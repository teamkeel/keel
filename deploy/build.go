package deploy

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
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

//go:embed lambdas/functions/dev-server.js
var devServer string

type BuildArgs struct {
	// Absolute path to Keel project
	ProjectRoot string
	Env         string
	// Where to pull the pre-built runtime binary from. Can be an absolute local path or a URL.
	RuntimeBinary string
	// A hook for modifying the schema after it's been loaded and validated. Our main use-case
	// for this is in integration tests where we make an API containing all models/actions.
	OnLoadSchema func(s *proto.Schema) *proto.Schema
}

type BuildResult struct {
	Schema *proto.Schema
	Config *config.ProjectConfig
	// Path to the directory containing the files needed for the runtime Lambda
	RuntimePath string
	// Path to the directory containing the files needed for the functions Lambda
	FunctionsPath string
}

func Build(ctx context.Context, args *BuildArgs) (*BuildResult, error) {
	heading(ctx, "Build")

	t := NewTiming()

	configFile, err := ResolveKeelConfig(ctx, &ResolveKeelConfigArgs{
		ProjectRoot: args.ProjectRoot,
		Env:         args.Env,
	})
	if err != nil {
		return nil, err
	}

	projectConfig := configFile.Config
	builder := schema.Builder{}
	protoSchema, err := builder.MakeFromDirectory(args.ProjectRoot)
	if err != nil {
		log(ctx, "%s your Keel schema contains errors. Run `keel validate` to see details on these errors", IconCross)
		return nil, err
	}
	if args.OnLoadSchema != nil {
		protoSchema = args.OnLoadSchema(protoSchema)
	}

	log(ctx, "%s Found %s schema file(s) %s", IconTick, orange("%d", len(builder.SchemaFiles())), t.Since())

	// No need to build the collector config file for local builds as we don't run it
	var collectorConfig *string
	if !isLocalBuild(args.Env) {
		collectorConfig, err = buildCollectorConfig(ctx, &BuildCollectorConfigArgs{
			ProjectRoot: args.ProjectRoot,
			Env:         args.Env,
			Config:      projectConfig,
		})
		if err != nil {
			return nil, err
		}
		if collectorConfig != nil {
			log(ctx, "%s Using OpenTelemetry collector config %s", IconTick, t.Since())
		}
	}

	runtimeResult, err := buildRuntime(ctx, &BuildRuntimeArgs{
		ProjectRoot:      args.ProjectRoot,
		Env:              args.Env,
		Schema:           protoSchema,
		Config:           projectConfig,
		RuntimeBinaryURL: args.RuntimeBinary,
		CollectorConfig:  collectorConfig,
	})
	if err != nil {
		return nil, err
	}
	relPath, err := filepath.Rel(args.ProjectRoot, runtimeResult.Path)
	if err != nil {
		return nil, err
	}
	log(ctx, "%s Built runtime into %s %s", IconTick, orange(relPath), t.Since())

	functionsResult, err := buildFunctions(ctx, &BuildFunctionsArgs{
		ProjectRoot:     args.ProjectRoot,
		Env:             args.Env,
		Schema:          protoSchema,
		Config:          projectConfig,
		CollectorConfig: collectorConfig,
	})
	if err != nil {
		return nil, err
	}

	return &BuildResult{
		Schema:        protoSchema,
		Config:        projectConfig,
		RuntimePath:   runtimeResult.Path,
		FunctionsPath: functionsResult.Path,
	}, nil
}

type BuildRuntimeArgs struct {
	// Absolute path of Keel project being built
	ProjectRoot string
	Env         string
	Schema      *proto.Schema
	Config      *config.ProjectConfig
	// Where to pull the pre-built runtime binary from. Can be a local path or URL.
	RuntimeBinaryURL string
	// A YAML string containing an OTEL collector config.
	CollectorConfig *string
}

type BuildRuntimeResult struct {
	SchemaPath string
	ConfigPath string
	Path       string
}

func buildRuntime(ctx context.Context, args *BuildRuntimeArgs) (*BuildRuntimeResult, error) {
	buildDir := filepath.Join(args.ProjectRoot, ".build")
	schemaPath := filepath.Join(buildDir, "runtime/schema.json")
	configPath := filepath.Join(buildDir, "runtime/config.json")

	err := os.MkdirAll(filepath.Join(buildDir, "runtime"), os.ModePerm)
	if err != nil {
		log(ctx, "%s error creating .build/runtime directory: %s", err.Error())
		return nil, err
	}

	schemaJson, err := protojson.Marshal(args.Schema)
	if err != nil {
		log(ctx, "%s error marshalling Keel schema to JSON", IconCross, err.Error())
		return nil, err
	}
	err = os.WriteFile(schemaPath, schemaJson, os.ModePerm)
	if err != nil {
		log(ctx, "%s error writing schema JSON to build directory: %s", IconCross, err.Error())
		return nil, err
	}

	// We use JSON here just so the runtime Lambda handler doesn't need to include a YAML parsing lib
	configJSON, err := json.Marshal(args.Config)
	if err != nil {
		log(ctx, "%s error marshalling Keel config to JSON", IconCross, err.Error())
		return nil, err
	}
	err = os.WriteFile(configPath, configJSON, os.ModePerm)
	if err != nil {
		log(ctx, "%s error writing schema JSON to build directory: %s", IconCross, err.Error())
		return nil, err
	}

	// No need to download the runtime binary for local builds as we don't run it
	if !isLocalBuild(args.Env) {
		var b []byte
		if strings.HasPrefix(args.RuntimeBinaryURL, "http") {
			t := NewTiming()
			res, err := http.Get(args.RuntimeBinaryURL)
			if err != nil {
				log(ctx, "%s Error requesting %s: %s", IconCross, args.RuntimeBinaryURL, err.Error())
				return nil, err
			}
			if res.StatusCode >= 300 {
				return nil, fmt.Errorf("non-200 (%d) trying to fetch runtime binary", res.StatusCode)
			}
			b, err = io.ReadAll(res.Body)
			if err != nil {
				log(ctx, "%s Error reading response from %s: %s", IconCross, args.RuntimeBinaryURL, err.Error())
				return nil, err
			}
			b, err = extractRuntimeBinary(b)
			if err != nil {
				log(ctx, "%s Error extracting runtime binary from archive %s: %s", IconCross, args.RuntimeBinaryURL, err.Error())
				return nil, err
			}
			log(ctx, "%s Fetched runtime binary from %s %s", IconTick, orange(args.RuntimeBinaryURL), t.Since())
		} else {
			p := args.RuntimeBinaryURL
			if !filepath.IsAbs(p) {
				p = filepath.Join(args.ProjectRoot, p)
			}
			b, err = os.ReadFile(p)
			if err != nil {
				log(ctx, "%s Error reading local runtime binary %s: %s", IconCross, p, err.Error())
				return nil, err
			}
		}

		err = os.WriteFile(filepath.Join(buildDir, "runtime/bootstrap"), b, os.ModePerm)
		if err != nil {
			log(ctx, "%s Error writing runtime binary to build directory: %s", IconCross, err.Error())
			return nil, err
		}
	}

	if args.CollectorConfig != nil {
		err = os.WriteFile(filepath.Join(buildDir, "runtime/collector.yaml"), []byte(*args.CollectorConfig), os.ModePerm)
		if err != nil {
			log(ctx, "%s Error writing collector config to build directory: %s", IconCross, err.Error())
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
	// Absolute path to Keel project being built
	ProjectRoot string
	Env         string
	Schema      *proto.Schema
	Config      *config.ProjectConfig
	// YAML string containing an OTEL collector config
	CollectorConfig *string
}

type BuildFunctionsResult struct {
	// Path to directory containing build for functions Lambda
	Path string
}

func buildFunctions(ctx context.Context, args *BuildFunctionsArgs) (*BuildFunctionsResult, error) {
	buildDir := filepath.Join(args.ProjectRoot, ".build")

	err := os.MkdirAll(filepath.Join(buildDir, "functions"), os.ModePerm)
	if err != nil {
		log(ctx, "%s error creating .build/functions directory: %s", err.Error())
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

	if isLocalBuild(args.Env) {
		sdk = append(sdk, &codegen.GeneratedFile{
			Path:     ".build/server.js",
			Contents: devServer,
		})
	}

	err = sdk.Write(args.ProjectRoot)
	if err != nil {
		return nil, err
	}

	deploy := args.Config.Deploy

	// Download the certificate for RDS as it requires SSL.
	// TODO: consider supporting a general "sslCert" option in config for external db's or adding explicit
	// support for services like Supabase/Neon and pulling the certs here manually.
	// For reference:
	// - Supabase cert: https://supabase-downloads.s3-ap-southeast-1.amazonaws.com/prod/ssl/prod-ca-2021.crt
	// - More info on Supabase SSL - https://supabase.com/docs/guides/platform/ssl-enforcement
	// - Neon cert: https://letsencrypt.org/certs/isrgrootx1.pem
	// - More info on Neon SSL - https://neon.tech/docs/connect/connect-securely
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
		log(ctx, "%s Downloaded RDS public key %s", IconTick, t.Since())
	}

	if args.CollectorConfig != nil {
		err = os.WriteFile(filepath.Join(buildDir, "functions/collector.yaml"), []byte(*args.CollectorConfig), os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	bundledPath := filepath.Join(buildDir, "functions/main-bundled.js")

	if !isLocalBuild(args.Env) {
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
				// TODO: do we need this
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
		log(ctx, "%s Built functions into %s %s", IconTick, orange(rel), t.Since())
	}

	return &BuildFunctionsResult{
		Path: filepath.Join(buildDir, "functions"),
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

	// Providing an OTEL collector file is optional, if not specified then bail
	telemetry := deploy.Telemetry
	if telemetry == nil || telemetry.Collector == "" {
		return nil, nil
	}

	collectorPath := filepath.Join(args.ProjectRoot, telemetry.Collector)

	// We allow the use of {{ .secrets.MY_SECRET }} in the collector which we will replace now, so first we need to fetch all secrets
	params, err := ListSecrets(ctx, &ListSecretsArgs{
		ProjectRoot: args.ProjectRoot,
		Env:         args.Env,
		Silent:      true,
	})
	if err != nil {
		return nil, err
	}

	secrets := lo.Reduce(params, func(m map[string]string, p types.Parameter, _ int) map[string]string {
		parts := strings.Split(*p.Name, "/")
		m[parts[len(parts)-1]] = *p.Value
		return m
	}, map[string]string{})

	templateName := filepath.Base(collectorPath)
	t := template.Must(template.New(templateName).Option("missingkey=error").ParseFiles(collectorPath))
	b := bytes.Buffer{}
	err = t.Execute(&b, map[string]any{
		"secrets": secrets,
	})
	if err != nil {
		log(ctx, "%s Error rendering secrets in OTEL collector config: %s", IconCross, gray(err.Error()))
		return nil, err
	}

	result := b.String()
	return &result, nil
}

func generateFunctionsHandler(schema *proto.Schema, cfg *config.ProjectConfig) (*codegen.GeneratedFile, error) {
	functions := map[string]string{}
	jobs := []string{}
	subscribers := []string{}
	flows := []string{}
	routes := []string{}
	actionTypes := map[string]string{}

	for _, model := range schema.GetModels() {
		for _, op := range model.GetActions() {
			if op.GetImplementation() != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}
			functions[op.GetName()] = op.GetName()
			actionTypes[op.GetName()] = op.GetType().String()
		}
	}

	if cfg != nil {
		for _, v := range cfg.Auth.EnabledHooks() {
			functions[string(v)] = fmt.Sprintf("auth/%s", string(v))
		}
	}

	for _, job := range schema.GetJobs() {
		jobName := strcase.ToLowerCamel(job.GetName())
		jobs = append(jobs, jobName)
	}

	for _, subscriber := range schema.GetSubscribers() {
		subscriberName := strcase.ToLowerCamel(subscriber.GetName())
		subscribers = append(subscribers, subscriberName)
	}

	for _, flow := range schema.GetFlows() {
		flowName := strcase.ToLowerCamel(flow.GetName())
		flows = append(flows, flowName)
	}

	for _, route := range schema.GetRoutes() {
		routes = append(routes, route.GetHandler())
	}

	var tmpl = template.Must(template.New("handler.js").Parse(functionsHandlerTemplate))

	b := bytes.Buffer{}
	err := tmpl.Execute(&b, map[string]interface{}{
		"Functions":   functions,
		"Subscribers": subscribers,
		"Jobs":        jobs,
		"Flows":       flows,
		"Routes":      routes,
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

func isLocalBuild(env string) bool {
	return env == "development" || env == "test"
}

func extractRuntimeBinary(b []byte) ([]byte, error) {
	gzr, err := gzip.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Name != "runtime-lambda" {
			continue
		}

		var out bytes.Buffer
		_, err = io.Copy(&out, tr)
		if err != nil {
			return nil, err
		}

		return out.Bytes(), nil
	}

	return nil, errors.New("runtime-lambda binary not found in archive")
}
