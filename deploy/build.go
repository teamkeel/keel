package deploy

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/iancoleman/strcase"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"google.golang.org/protobuf/encoding/protojson"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

//go:embed lambdas/functions/main.js.go.tmpl
var functionsHandlerTemplate string

type BuildArgs struct {
	ProjectRoot       string
	Env               string
	RuntimeBinary     string
	SkipRuntimeBinary bool
	OnLoadSchema      func(s *proto.Schema) *proto.Schema
}

type BuildResult struct {
	Schema               *proto.Schema
	SchemaPath           string
	Config               *config.ProjectConfig
	ConfigPath           string
	RuntimePath          string
	FunctionsBundledPath string
	FunctionsPath        string
}

func Build(ctx context.Context, args *BuildArgs) (*BuildResult, error) {
	heading("Build")

	t := NewTiming()

	projectConfig, err := loadKeelConfig(&LoadKeelConfigArgs{
		ProjectRoot: args.ProjectRoot,
		Env:         args.Env,
	})
	if err != nil {
		return nil, err
	}

	builder := schema.Builder{}
	protoSchema, err := builder.MakeFromDirectory(args.ProjectRoot)
	if err != nil {
		log("%s your Keel schema contains errors. Run `keel validate` to see details on these errors", IconCross)
		return nil, err
	}
	if args.OnLoadSchema != nil {
		protoSchema = args.OnLoadSchema(protoSchema)
	}

	log("%s Found %s schema file(s) %s", IconTick, orange("%d", len(builder.SchemaFiles())), t.Since())

	collectorConfig, err := buildCollectorConfig(ctx, &BuildCollectorConfigArgs{
		ProjectRoot: args.ProjectRoot,
		Env:         args.Env,
		Config:      projectConfig,
	})
	if err != nil {
		return nil, err
	}
	if collectorConfig != nil {
		log("%s Using OpenTelemetry collector config", IconTick, orange("%d", len(builder.SchemaFiles())), t.Since())
	}

	runtimeResult, err := buildRuntime(ctx, &BuildRuntimeArgs{
		ProjectRoot:               args.ProjectRoot,
		Env:                       args.Env,
		Schema:                    protoSchema,
		Config:                    projectConfig,
		RuntimeBinaryURL:          args.RuntimeBinary,
		SkipRuntimeBinaryDownload: args.SkipRuntimeBinary,
		CollectorConfig:           collectorConfig,
	})
	if err != nil {
		return nil, err
	}
	relPath, err := filepath.Rel(args.ProjectRoot, runtimeResult.Path)
	if err != nil {
		return nil, err
	}
	log("%s Built runtime into %s %s", IconTick, orange(relPath), t.Since())

	functionsResult, err := buildFunctions(ctx, &BuildFunctionsArgs{
		ProjectRoot:     args.ProjectRoot,
		Schema:          protoSchema,
		Config:          projectConfig,
		CollectorConfig: collectorConfig,
	})
	if err != nil {
		return nil, err
	}

	return &BuildResult{
		Schema:               protoSchema,
		SchemaPath:           runtimeResult.SchemaPath,
		Config:               projectConfig,
		ConfigPath:           runtimeResult.ConfigPath,
		RuntimePath:          runtimeResult.Path,
		FunctionsBundledPath: functionsResult.BundledPath,
		FunctionsPath:        functionsResult.Path,
	}, nil
}

type BuildRuntimeArgs struct {
	ProjectRoot               string
	Env                       string
	Schema                    *proto.Schema
	Config                    *config.ProjectConfig
	RuntimeBinaryURL          string
	SkipRuntimeBinaryDownload bool
	CollectorConfig           *string
}

type BuildRuntimeResult struct {
	SchemaPath string
	ConfigPath string
	Path       string
}

func buildRuntime(
	ctx context.Context, // nolint: unparam
	args *BuildRuntimeArgs,
) (*BuildRuntimeResult, error) {
	buildDir := filepath.Join(args.ProjectRoot, ".build")
	schemaPath := filepath.Join(buildDir, "runtime/schema.json")
	configPath := filepath.Join(buildDir, "runtime/config.json")

	err := os.MkdirAll(filepath.Join(buildDir, "runtime"), os.ModePerm)
	if err != nil {
		log("%s error creating .build/runtime directory: %s", err.Error())
		return nil, err
	}

	schemaJson, err := protojson.Marshal(args.Schema)
	if err != nil {
		log("%s error marshalling Keel schema to JSON", IconCross, err.Error())
		return nil, err
	}
	err = os.WriteFile(schemaPath, schemaJson, os.ModePerm)
	if err != nil {
		log("%s error writing schema JSON to build directory: %s", IconCross, err.Error())
		return nil, err
	}

	configJSON, err := json.Marshal(args.Config)
	if err != nil {
		log("%s error marshalling Keel config to JSON", IconCross, err.Error())
		return nil, err
	}
	err = os.WriteFile(configPath, configJSON, os.ModePerm)
	if err != nil {
		log("%s error writing schema JSON to build directory: %s", IconCross, err.Error())
		return nil, err
	}

	if !args.SkipRuntimeBinaryDownload {
		var b []byte
		if strings.HasPrefix(args.RuntimeBinaryURL, "http") {
			t := NewTiming()
			res, err := http.Get(args.RuntimeBinaryURL)
			if err != nil {
				log("%s error requesting %s: %s", IconCross, args.RuntimeBinaryURL, err.Error())
				return nil, err
			}
			if res.StatusCode >= 300 {
				return nil, fmt.Errorf("non-200 (%d) trying to fetch runtime binary", res.StatusCode)
			}
			b, err = io.ReadAll(res.Body)
			if err != nil {
				log("%s reading response from %s: %s", IconCross, args.RuntimeBinaryURL, err.Error())
				return nil, err
			}
			log("%s Fetched runtime binary from %s %s", IconTick, orange(args.RuntimeBinaryURL), t.Since())
		} else {
			p := args.RuntimeBinaryURL
			if !filepath.IsAbs(p) {
				p = filepath.Join(args.ProjectRoot, p)
			}
			b, err = os.ReadFile(p)
			if err != nil {
				log("%s error reading local runtime binary %s: %s", IconCross, p, err.Error())
				return nil, err
			}
		}

		err = os.WriteFile(filepath.Join(buildDir, "runtime/bootstrap"), b, os.ModePerm)
		if err != nil {
			log("%s error writing runtime binary to build directory: %s", IconCross, err.Error())
			return nil, err
		}
	}

	if args.CollectorConfig != nil {
		err = os.WriteFile(filepath.Join(buildDir, "runtime/collector.yaml"), []byte(*args.CollectorConfig), os.ModePerm)
		if err != nil {
			log("%s error writing collector config to build directory: %s", IconCross, err.Error())
			return nil, err
		}
	}

	return &BuildRuntimeResult{
		SchemaPath: schemaPath,
		ConfigPath: configPath,
		Path:       filepath.Join(buildDir, "runtime"),
	}, nil
}

type BuildFunctionsArgs struct {
	ProjectRoot     string
	Schema          *proto.Schema
	Config          *config.ProjectConfig
	CollectorConfig *string
}

type BuildFunctionsResult struct {
	BundledPath string
	Path        string
}

func buildFunctions(ctx context.Context, args *BuildFunctionsArgs) (*BuildFunctionsResult, error) {
	buildDir := filepath.Join(args.ProjectRoot, ".build")

	err := os.MkdirAll(filepath.Join(buildDir, "functions"), os.ModePerm)
	if err != nil {
		log("%s error creating .build/functions directory: %s", err.Error())
		return nil, err
	}

	sdk, err := node.Generate(ctx, args.Schema, args.Config)
	if err != nil {
		return nil, err
	}

	functionsHandler, err := generateFunctionsHandler(args.Schema, args.Config)
	if err != nil {
		return nil, err
	}

	sdk = append(sdk, functionsHandler)

	err = sdk.Write(args.ProjectRoot)
	if err != nil {
		return nil, err
	}

	deploy := args.Config.Deploy
	if deploy != nil && (deploy.Database == nil || deploy.Database.Provider == "rds") {
		t := NewTiming()
		res, err := http.Get("https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem")
		if err != nil {
			return nil, err
		}
		if res.StatusCode >= 300 {
			return nil, fmt.Errorf("non-200 (%d) fetching .pem for RDS", res.StatusCode)
		}
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(filepath.Join(buildDir, "functions/rds.pem"), b, os.ModePerm)
		if err != nil {
			return nil, err
		}
		log("%s Downloaded RDS public key %s", IconTick, t.Since())
	}

	if args.CollectorConfig != nil {
		err = os.WriteFile(filepath.Join(buildDir, "functions/collector.yaml"), []byte(*args.CollectorConfig), os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	bundledPath := filepath.Join(buildDir, "functions/main-bundled.js")

	t := NewTiming()
	res := esbuild.Build(esbuild.BuildOptions{
		EntryPoints:    []string{filepath.Join(buildDir, "functions/main.js")},
		Outfile:        bundledPath,
		Bundle:         true,
		Write:          true,
		AllowOverwrite: true,
		Target:         esbuild.ESNext,
		Platform:       esbuild.PlatformNode,
		Format:         esbuild.FormatCommonJS,
		Sourcemap:      esbuild.SourceMapLinked,
		External:       []string{"pg-native"},
		Loader: map[string]esbuild.Loader{
			".node": esbuild.LoaderFile,
		},
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		KeepNames:         true,
	})
	if len(res.Errors) > 0 {
		return nil, fmt.Errorf("esbuild error: %s", res.Errors[0].Text)
	}

	rel, _ := filepath.Rel(args.ProjectRoot, filepath.Join(buildDir, "functions"))
	log("%s Built functions into %s %s", IconTick, orange(rel), t.Since())

	return &BuildFunctionsResult{
		BundledPath: bundledPath,
		Path:        filepath.Join(buildDir, "functions"),
	}, nil
}

type BuildCollectorConfigArgs struct {
	ProjectRoot string
	Env         string
	Config      *config.ProjectConfig
}

func buildCollectorConfig(ctx context.Context, args *BuildCollectorConfigArgs) (*string, error) {
	deploy := args.Config.Deploy
	if deploy == nil {
		return nil, nil
	}

	telemetry := deploy.Telemetry
	if telemetry == nil || telemetry.Collector == "" {
		return nil, nil
	}

	collectorPath := filepath.Join(args.ProjectRoot, telemetry.Collector)

	params, err := ListSecrets(ctx, &ListSecretsArgs{
		ProjectRoot: args.ProjectRoot,
		Env:         args.Env,
	})
	if err != nil {
		return nil, err
	}

	secrets := lo.Reduce(params, func(m map[string]string, p types.Parameter, _ int) map[string]string {
		m[*p.Name] = *p.Value
		return m
	}, map[string]string{})

	t := template.Must(template.New("collector").ParseFiles(collectorPath))
	b := bytes.Buffer{}
	err = t.Execute(&b, map[string]any{
		"secrets": secrets,
	})
	if err != nil {
		log("%s error rendering secrets in OTEL collector config: %s", IconCross, err.Error())
		return nil, err
	}

	result := b.String()
	return &result, nil
}

func generateFunctionsHandler(schema *proto.Schema, cfg *config.ProjectConfig) (*codegen.GeneratedFile, error) {
	functions := map[string]string{}
	jobs := []string{}
	subscribers := []string{}
	actionTypes := map[string]string{}

	for _, model := range schema.Models {
		for _, op := range model.Actions {
			if op.Implementation != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}
			functions[op.Name] = op.Name
			actionTypes[op.Name] = op.Type.String()
		}
	}

	if cfg != nil {
		for _, v := range cfg.Auth.EnabledHooks() {
			functions[string(v)] = fmt.Sprintf("auth/%s", string(v))
		}
	}

	for _, job := range schema.Jobs {
		jobName := strcase.ToLowerCamel(job.Name)
		jobs = append(jobs, jobName)
	}

	for _, subscriber := range schema.Subscribers {
		subscriberName := strcase.ToLowerCamel(subscriber.Name)
		subscribers = append(subscribers, subscriberName)
	}

	var tmpl = template.Must(template.New("handler.js").Parse(functionsHandlerTemplate))

	b := bytes.Buffer{}
	err := tmpl.Execute(&b, map[string]interface{}{
		"Functions":   functions,
		"Subscribers": subscribers,
		"Jobs":        jobs,
		"ActionTypes": actionTypes,
	})
	if err != nil {
		return nil, err
	}

	return &codegen.GeneratedFile{
		Path:     ".build/functions/main.js",
		Contents: b.String(),
	}, nil
}
