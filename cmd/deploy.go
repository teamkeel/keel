package cmd

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Builds and deploys your app ready to your AWS account",
	Run: func(cmd *cobra.Command, args []string) {
		result := build()
		if result == nil {
			return
		}

		fmt.Println("")
		ok := deploy(result)
		if !ok {
			return
		}

		fmt.Println("All done 🎉")
	},
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds your app ready for deploying to your AWS account",
	Run: func(cmd *cobra.Command, args []string) {
		_ = build()
	},
}

func deploy(buildResult *BuildResult) bool {
	fmt.Println(colors.White("Deploying your app to AWS:").Highlight().String())
	fmt.Println("")

	ctx := context.Background()

	printUpdate("Checking AWS auth")

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		printError("Error loading AWS config", err)
		return false
	}

	_, err = sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		printError("Error checking AWS auth", err)
		return false
	}

	printUpdate("Checking private key")

	ssmClient := ssm.NewFromConfig(cfg)

	privateKeyParamName := fmt.Sprintf("/sst/%s/%s/Secret/KEEL_PRIVATE_KEY/value", buildResult.Config.Project.Name, "prod")

	getParamResult, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
		Name: aws.String(privateKeyParamName),
	})
	if err != nil {
		var ae smithy.APIError
		if !errors.As(err, &ae) || ae.ErrorCode() != "ParameterNotFound" {
			printError("Error checking for private key secret", err)
			return false
		}
	}

	if getParamResult == nil || getParamResult.Parameter == nil {
		printUpdate("Generating private key secret")
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			printError("Error generating private key", err)
			return false
		}

		privateKeyPem := pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		})

		_, err = ssmClient.PutParameter(ctx, &ssm.PutParameterInput{
			Name:  aws.String(privateKeyParamName),
			Value: aws.String(string(privateKeyPem)),
			Type:  ssmtypes.ParameterTypeSecureString,
		})
		if err != nil {
			printError("Error creating private key secret", err)
			return false
		}
	}

	if buildResult.Config.Deploy.Database.Provider == "external" {
		printUpdate("Checking for external database url secret")
		_, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
			Name: aws.String(fmt.Sprintf("/sst/%s/%s/Secret/POSTGRES_URL/value", buildResult.Config.Project.Name, "prod")),
		})
		if err != nil {
			printError("Database config is set to 'external' but the POSTGRES_URL secret does not seem to be set - this is required when deploying using an external database", err)
			fmt.Println("")
			fmt.Println("To set the POSTGRES_URL secret run:\n\n  keel secrets set <env> POSTGRES_URL 'url-of-your-postgres-database'")
			return false
		}
	}

	printUpdate("Deploying with SST")
	fmt.Println("")

	sst := exec.Command("../node_modules/.bin/sst", "deploy", "--stage", "prod")
	sst.Dir = ".build"
	sst.Stdout = os.Stdout
	sst.Stderr = os.Stderr
	return sst.Run() != nil
}

type BuildResult struct {
	Config *config.ProjectConfig
	Schema *proto.Schema
}

func build() *BuildResult {
	fmt.Println(colors.White("Building your Keel app for deploying to AWS:").Highlight().String())
	fmt.Println("")

	builder := &schema.Builder{}
	protoSchema, err := builder.MakeFromDirectory(".")
	if err != nil {
		printError("Error loading Keel project", err)
		return nil
	}

	if builder.Config.Project.Name == "" {
		printError("Please set a project.name in keelconfig.yaml", nil)
		return nil
	}

	printUpdate("Cleaning build directory")
	err = os.RemoveAll(".build")
	if err != nil {
		printError("Error removing build directory", err)
		return nil
	}

	err = os.MkdirAll(".build", 0777)
	if err != nil {
		printError("Error creating build directory", err)
		return nil
	}

	var steps = []func(c *config.ProjectConfig, schema *proto.Schema) error{
		installNpmDeps,
		tsConfig,
		writeSchema,
		sstConfig,
		sstStacks,
		generateLambdas,
	}

	for _, step := range steps {
		err = step(builder.Config, protoSchema)
		if err != nil {
			return nil
		}
	}

	printUpdate("Compiling TypeScript")
	cmd := exec.Command("./node_modules/.bin/tsc", "--noEmit")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		printError("TypeScript compile error", err)
		return nil
	}

	printUpdate("Building infra")
	cmd = exec.Command("../node_modules/.bin/sst", "build", "--stage", "prod")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = ".build"
	err = cmd.Run()
	if err != nil {
		return nil
	}

	fmt.Println("")
	printSuccess("Build complete")

	return &BuildResult{
		Config: builder.Config,
		Schema: protoSchema,
	}
}

func generateLambdas(config *config.ProjectConfig, schema *proto.Schema) error {
	printUpdate("Generating Lambdas")

	// This is currently just set up for testing locally. For development we probably need a flag
	// for passing the path to the keel directory, and then for published CLI binaries we need
	// to just require the same version as the CLI
	gomod := `
module %s

go 1.20

replace github.com/teamkeel/keel => ../../keel

require github.com/teamkeel/keel v0.0.0-00010101000000-000000000000
`

	gomod = fmt.Sprintf(gomod, config.Project.Name)
	err := writeFile(".build/go.mod", gomod)
	if err != nil {
		return err
	}

	handler := `
package main

import (
	"github.com/teamkeel/keel/infra/runtime"
)

func main() {
	runtime.Start()
}
`

	err = writeFile(".build/lambdas/runtime/main.go", handler)
	if err != nil {
		return err
	}

	handler = `
	package main
	
	import (
		"github.com/teamkeel/keel/infra/migrations"
	)
	
	func main() {
		migrations.Start()
	}
	`

	err = writeFile(".build/lambdas/migrations/main.go", handler)
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = ".build"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return generateFunctionsLambda(schema, config)
}

func generateFunctionsLambda(schema *proto.Schema, config *config.ProjectConfig) error {
	if !node.HasFunctions(schema, config) {
		return nil
	}

	content := `
import { SecretsManagerClient, GetSecretValueCommand } from "@aws-sdk/client-secrets-manager";
import opentelemetry from "@opentelemetry/api";
import { Config } from "sst/node/config";
import { RDS } from "sst/node/rds";

function once(fn) {
	let value;
	return function() {
		if (!value) {
			value = fn(...arguments);
		}
		return value;
	};
}

const getFunctionsMapping = once(async () => {
`

	functionNames := []string{}
	imports := []string{}
	actionTypes := map[string]string{}

	for _, model := range schema.Models {
		for _, op := range model.Actions {
			if op.Implementation != proto.ActionImplementation_ACTION_IMPLEMENTATION_CUSTOM {
				continue
			}
			functionNames = append(functionNames, op.Name)
			imports = append(imports, fmt.Sprintf(`  const function_%s = await import("../../../functions/%s");`, op.Name, op.Name))
			actionTypes[op.Name] = op.Type.String()
		}
	}

	content += strings.Join(imports, "\n") + "\n\n"

	content += "  const functions = {\n"
	for _, name := range functionNames {
		content += fmt.Sprintf("    %s: function_%s.default,\n", name, name)
	}
	content += "  };\n"

	content += "  const actionTypes = {\n"
	for functionName, actionType := range actionTypes {
		content += fmt.Sprintf("    %s: \"%s\",\n", functionName, actionType)
	}
	content += "  };\n"

	content += `

  return { functions, actionTypes };
});

async function setDatabaseEnvVar() {
	if (process.env.KEEL_DB_CONN) {
		return;
	}

	try {
		process.env.KEEL_DB_CONN = Config.POSTGRES_URL;
	} catch (err) {
		// this will error if the POSTGRES_URL secret is not bound to the function
	}

	if (process.env.KEEL_DB_CONN) {
		return;
	}

	let secretArn = '';

	try {
		secretArn = RDS.Database.secretArn;
	} catch (err) {
		// this will error if the RDS database resource is not bound to the function
	}

	if (!secretArn) {
		throw new Error("neither POSTGRES_URL secret or RDS database details bound to function");
	}

	const client = new SecretsManagerClient({ region: process.env.SST_REGION });

	const command = new GetSecretValueCommand({
		SecretId: secretArn,
	});

	const res = await client.send(command);

	const creds = JSON.parse(res.SecretString);

	process.env.KEEL_DB_CONN = ` + "`postgres://${creds.username}:${creds.password}@${creds.host}/${creds.dbname}`;" + `
}

export const handler = async (event) => {
	await setDatabaseEnvVar()

	// These need to come after setting the db env var
	const { handleRequest } = await import("@teamkeel/functions-runtime");
	const { createContextAPI, permissionFns } = await import("@teamkeel/sdk");
	const { functions, actionTypes } = await getFunctionsMapping();

	if (event.rawPath === "/_health") {
		return {
			id: "ok",
			result: {},
		};
	}

	let rpcResponse = null;
	switch (event.type) {
	case "action":
		rpcResponse = await handleRequest(event, {
			functions,
			createContextAPI,
			actionTypes,
			permissionFns,
		});
		break;
	}

	// The "delegate" is the actual provider set by the functions-runtime package
	const provider = opentelemetry.trace.getTracerProvider().getDelegate();
	if (provider && provider.forceFlush) {
		await provider.forceFlush();
	}

	return rpcResponse;
};`

	return writeFile(".build/lambdas/functions/index.js", content)
}

const stacksIndexTemplate = `
import {
  StackContext,
  Function,
  Bucket,
  RDS,
  Script,
  FunctionProps,
  Config,
} from "sst/constructs";
import { SSTConstruct } from "sst/constructs/Construct";
import { BucketDeployment, Source } from "aws-cdk-lib/aws-s3-deployment";
import { Vpc, SubnetType } from "aws-cdk-lib/aws-ec2";
import { RemovalPolicy } from "aws-cdk-lib/core";

type MainStackOptions = {
  secrets: Array<string>;
  externalDatabase: boolean;
  schemaFileKey: string;
  hasFunctions: boolean;
};

function mainStack({ stack }: StackContext, options: MainStackOptions) {
  const bucket = new Bucket(stack, "RuntimeAssets", {
    cdk: {
      bucket: {
        removalPolicy: RemovalPolicy.DESTROY,
		autoDeleteObjects: true,
      },
    },
  });

  const assets = new BucketDeployment(stack, "ProtoSchema", {
    sources: [Source.asset("./assets")],
    destinationBucket: bucket.cdk.bucket,
  });

  const functionBindings: Array<SSTConstruct> = [bucket];

  for (const secret of options.secrets) {
    const s = new Config.Secret(stack, secret);
    functionBindings.push(s);
  }

  let vpc: Vpc | undefined;
  let rds: RDS | undefined;

  if (!options.externalDatabase) {
    vpc = new Vpc(stack, "Vpc", {
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: "ingress",
          subnetType: SubnetType.PUBLIC,
        },
        {
          cidrMask: 24,
          name: "application",
          subnetType: SubnetType.PRIVATE_WITH_EGRESS,
        },
        {
          cidrMask: 28,
          name: "rds",
          subnetType: SubnetType.PRIVATE_ISOLATED,
        },
      ],
    });

    rds = new RDS(stack, "Database", {
      engine: "postgresql13.9",
      defaultDatabaseName: "keel",
      cdk: {
        cluster: {
          vpc,
          vpcSubnets: {
            subnetType: SubnetType.PRIVATE_ISOLATED,
          },
        },
      },
    });

    rds.cdk.cluster.connections.allowDefaultPortFromAnyIpv4();

    if (rds) {
      functionBindings.push(rds);
    }
  }

  const funcDefaults: Partial<FunctionProps> = {
    runtime: "go",
	timeout: "60 seconds",
    bind: functionBindings,
    environment: {
      KEEL_SCHEMA_FILE_KEY: options.schemaFileKey,
	  KEEL_DB_CONN_TYPE: "pg",
    },
    vpc,
    vpcSubnets: vpc
      ? {
        subnetType: SubnetType.PRIVATE_WITH_EGRESS,
      }
      : undefined,
  };

  const migrationsFunc = new Function(stack, "MigrationsHandler", {
    ...funcDefaults,
    handler: "lambdas/migrations/main.go",
    timeout: "5 minute",
  });

  const migrationsScript = new Script(stack, "Migrations", {
    onCreate: migrationsFunc,
    onUpdate: migrationsFunc,
  });
  migrationsScript.node.addDependency(assets);
  if (rds) {
    migrationsScript.node.addDependency(rds);
  }

  const runtime = new Function(stack, "RuntimeHandler", {
    ...funcDefaults,
    handler: "lambdas/runtime/main.go",
    url: true,
  });

  if (options.hasFunctions) {
    const functions = new Function(stack, "FunctionsHandler", {
      ...funcDefaults,
      runtime: "nodejs18.x",
      handler: "lambdas/functions/index.handler",
    });

    runtime.bind([functions]);
  }

  stack.addOutputs({
    ApiEndpoint: runtime.url,
  });
}

export function MainStack(ctx: StackContext) {
  mainStack(ctx, {
    secrets: [ {{ range .Secrets }} "{{ . }}", {{ end }}],
    externalDatabase: {{ .ExternalDatabase }},
    schemaFileKey: "{{ .SchemaFileKey }}",
	hasFunctions: {{ .HasFunctions }},
  });
}

	`

type StacksTemplateData struct {
	Secrets          []string
	ExternalDatabase bool
	SchemaFileKey    string
	HasFunctions     bool
}

func sstStacks(config *config.ProjectConfig, schema *proto.Schema) error {
	printUpdate("Generating infrasture code")

	t := template.Must(template.New("stacks/index.ts").Parse(stacksIndexTemplate))

	entries, err := os.ReadDir(".build/assets/schema")
	if err != nil {
		return err
	}

	secretNames := []string{"KEEL_PRIVATE_KEY"}
	externalDatabase := false
	if config.Deploy.Database.Provider == "external" {
		secretNames = append(secretNames, "POSTGRES_URL")
		externalDatabase = true
	}

	var b bytes.Buffer
	err = t.Execute(&b, &StacksTemplateData{
		Secrets:          secretNames,
		ExternalDatabase: externalDatabase,
		SchemaFileKey:    "schema/" + entries[0].Name(),
		HasFunctions:     node.HasFunctions(schema, config),
	})
	if err != nil {
		printError("Error generating infrastructure code", err)
		return err
	}

	return writeFile(".build/stacks/index.ts", b.String())
}

func sstConfig(config *config.ProjectConfig, schema *proto.Schema) error {
	c := `
// Generated
import { SSTConfig } from "sst";
import { MainStack } from "./stacks";

export default {
  config(_input) {
    return {
      name: "%s",
      region: "eu-west-2",
    };
  },
  stacks(app) {
    app.stack(MainStack);
  }
} satisfies SSTConfig;`

	c = fmt.Sprintf(c, config.Project.Name)
	return writeFile(".build/sst.config.ts", c)
}

func writeSchema(c *config.ProjectConfig, s *proto.Schema) error {
	b, err := protojson.Marshal(s)
	if err != nil {
		return err
	}

	hash := sha1.New()
	_, _ = hash.Write(b)
	sha := hex.EncodeToString(hash.Sum(nil))
	filename := fmt.Sprintf(".build/assets/schema/schema.%s.json", sha)

	return writeFile(filename, string(b))
}

type PackageJson struct {
	Name            string         `json:"name"`
	Version         string         `json:"version"`
	DevDependencies map[string]any `json:"devDependencies"`
}

var sstNpmDeps = map[string]any{
	"sst":                             "2.24.14",
	"@aws-sdk/client-secrets-manager": "3.398.0",
	"aws-cdk-lib":                     "2.91.0",
	"constructs":                      "10.2.69",
	"typescript":                      "5.2.2",
	"@tsconfig/node16":                "16.1.1",
	"@types/node":                     "20.5.6",
}

func installNpmDeps(config *config.ProjectConfig, schema *proto.Schema) error {
	printUpdate("Checking JS dependencies")

	// If schema has functions do a normal bootstrap
	if node.HasFunctions(schema, config) {
		err := node.Bootstrap(".")
		if err != nil {
			return err
		}
	}

	_, err := os.Stat("package.json")
	var deps map[string]any
	if errors.Is(err, os.ErrNotExist) {
		p := PackageJson{
			Name:            config.Project.Name,
			Version:         "0.0.0",
			DevDependencies: sstNpmDeps,
		}
		b, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			return err
		}

		err = writeFile("package.json", string(b))
		if err != nil {
			return err
		}
	} else {
		b, err := os.ReadFile("package.json")
		if err != nil {
			return err
		}

		var p PackageJson
		err = json.Unmarshal(b, &p)
		if err != nil {
			return err
		}

		deps = p.DevDependencies
	}

	toInstall := []string{}
	for k, v := range sstNpmDeps {
		installed, ok := deps[k]
		if !ok || installed != v {
			toInstall = append(toInstall, fmt.Sprintf("%s@%s", k, v))
		}
	}

	if len(toInstall) > 0 {
		// TODO: break out dev and non-dev - I think only sst is a non-dev (as it's used in the functions lambda)
		args := append([]string{"install", "--save"}, toInstall...)
		c := exec.Command("npm", args...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err = c.Run()
		if err != nil {
			return err
		}
	} else {
		_, err = os.Stat("node_modules")
		if err != nil {
			c := exec.Command("npm", "install")
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			err = c.Run()
			if err != nil {
				return err
			}
		}
	}

	if node.HasFunctions(schema, config) {
		files, err := node.Generate(context.Background(), schema, config)
		if err != nil {
			return err
		}

		return files.Write(".")
	}

	return nil
}

func tsConfig(config *config.ProjectConfig, schema *proto.Schema) error {
	printUpdate("Checking TypeScript config")

	defaultTsConfig := `{
  "compilerOptions": {
    "lib": ["ES2016"],
    "target": "ES2016",
    "esModuleInterop": true,
    "moduleResolution": "node",
    "skipLibCheck": true,
    "strictNullChecks": true,
    "types": ["node"],
    "allowJs": true,
    "resolveJsonModule": true
  },
  "include": ["**/*.ts", ".build/stacks/index.ts"],
  "exclude": ["node_modules"]
}
`

	_, err := os.Stat("tsconfig.json")
	if errors.Is(err, os.ErrNotExist) {
		return writeFile("tsconfig.json", defaultTsConfig)
	}

	b, err := os.ReadFile("tsconfig.json")
	if err != nil {
		printError("Error reading existing tsconfig", err)
		return err
	}

	var tsconfig map[string]any
	err = json.Unmarshal(b, &tsconfig)
	if err != nil {
		return err
	}

	tsconfig["include"] = []string{"**/*.ts", ".build/stacks/index.ts"}

	b, err = json.MarshalIndent(tsconfig, "", "    ")
	if err != nil {
		return err
	}

	return writeFile("tsconfig.json", string(b))
}

func printUpdate(msg string) {
	bar := colors.Yellow("|").Highlight().String()
	msg = colors.Gray(msg).String()
	fmt.Println(bar, msg)
}

func printSuccess(msg string) {
	bar := colors.Green("✔").String()
	msg = colors.Gray(msg).String()
	fmt.Println(bar, msg)
}

func printError(msg string, err error) {
	bar := colors.Red("✗").String()
	msg = colors.White(msg + ": ").Highlight().String()
	if err != nil {
		msg += colors.Gray(err.Error()).String()
	}
	fmt.Println(bar, msg)
}

func writeFile(path string, contents string) error {
	d := filepath.Dir(path)
	if d != "." {
		err := os.MkdirAll(filepath.Dir(path), 0777)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(path, []byte(contents), 0777)
}

func init() {
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(deployCmd)
}
