package deploy

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/goccy/go-yaml"
	"github.com/iancoleman/strcase"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/teamkeel/keel/codegen"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"google.golang.org/protobuf/encoding/protojson"

	esbuild "github.com/evanw/esbuild/pkg/api"
)

//go:embed lambdas/functions/main.js.go.tmpl
var functionsHandlerTemplate string

type BuildLambdasArgs struct {
	Config           *config.ProjectConfig
	Schema           *proto.Schema
	RuntimeBinaryURL string
	ProjectRoot      string
	Events           chan Output
}

type BuildLambdasResult struct {
	RuntimeLambda   pulumi.Archive
	FunctionsLambda pulumi.Archive
}

func buildLambdas(ctx context.Context, args *BuildLambdasArgs) (*BuildLambdasResult, error) {
	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "Cleaning .build directory",
	}

	buildDir := filepath.Join(args.ProjectRoot, ".build")
	err := os.RemoveAll(buildDir)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error cleaning build directory",
			Error:   err,
		}
		return nil, err
	}

	schemaJson, err := protojson.Marshal(args.Schema)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error marshalling proto schema to JSON",
			Error:   err,
		}
		return nil, err
	}

	configYaml, err := yaml.Marshal(args.Config)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error marshalling config to YAML",
			Error:   err,
		}
		return nil, err
	}

	sdk, err := node.Generate(ctx, args.Schema, args.Config)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error generating functions SDK",
			Error:   err,
		}
		return nil, err
	}
	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "Generated functions SDK",
	}

	functionsHandler, err := generateFunctionsHandler(args.Schema, args.Config)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error generating functions handler",
			Error:   err,
		}
		return nil, err
	}
	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "Generated functions lambda handler",
	}

	sdk = append(sdk, functionsHandler)

	err = sdk.Write(args.ProjectRoot)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error writing generated code to build directory",
			Error:   err,
		}
		return nil, err
	}

	res := esbuild.Build(esbuild.BuildOptions{
		EntryPoints:    []string{filepath.Join(buildDir, "main.js")},
		Outdir:         buildDir,
		Write:          true,
		AllowOverwrite: true,
		Target:         esbuild.ESNext,
		Platform:       esbuild.PlatformNode,
		Format:         esbuild.FormatCommonJS,
		Sourcemap:      esbuild.SourceMapLinked,
		Bundle:         true,
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
		err = fmt.Errorf("esbuild error: %s", res.Errors[0].Text)
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error building functions",
			Error:   err,
		}
		return nil, err
	}
	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "Built functions handler",
	}

	functionsAssets := map[string]interface{}{
		"rds.pem": pulumi.NewRemoteAsset("https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem"),
	}

	for _, f := range res.OutputFiles {
		filename := filepath.Base(f.Path)
		functionsAssets[filename] = pulumi.NewFileAsset(f.Path)
	}

	var bootstrap pulumi.Asset
	if strings.HasPrefix(args.RuntimeBinaryURL, "http") {
		bootstrap = pulumi.NewRemoteAsset(args.RuntimeBinaryURL)
	} else {
		abspath, err := filepath.Abs(args.RuntimeBinaryURL)
		if err != nil {
			return nil, err
		}
		bootstrap = pulumi.NewFileAsset(abspath)
	}

	return &BuildLambdasResult{
		RuntimeLambda: pulumi.NewAssetArchive(map[string]interface{}{
			"bootstrap":       bootstrap,
			"schema.json":     pulumi.NewStringAsset(string(schemaJson)),
			"keelconfig.yaml": pulumi.NewStringAsset(string(configYaml)),
		}),
		FunctionsLambda: pulumi.NewAssetArchive(functionsAssets),
	}, nil
}

func generateFunctionsHandler(schema *proto.Schema, cfg *config.ProjectConfig) (*codegen.GeneratedFile, error) {
	functions := []string{}
	jobs := []string{}
	subscribers := []string{}
	actionTypes := map[string]string{}

	for _, model := range schema.Models {
		for _, op := range model.Actions {
			if op.Implementation != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}
			functions = append(functions, op.Name)
			actionTypes[op.Name] = op.Type.String()
		}
	}

	if cfg != nil {
		for _, v := range cfg.Auth.EnabledHooks() {
			functions = append(functions, string(v))
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
		Path:     ".build/main.js",
		Contents: b.String(),
	}, nil
}
