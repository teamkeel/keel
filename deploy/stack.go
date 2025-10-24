package deploy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
)

type SelectStackArgs struct {
	AwsConfig           aws.Config
	PulumiConfig        *PulumiConfig
	Env                 string
	Schema              *proto.Schema
	Config              *config.ProjectConfig
	AccountID           string
	RuntimeLambdaPath   string
	FunctionsLambdaPath string
}

func selectStack(ctx context.Context, args *SelectStackArgs) (*auto.Stack, error) {
	runFunc := createProgram(&NewProgramArgs{
		AwsConfig:           args.AwsConfig,
		AwsAccountID:        args.AccountID,
		RuntimeLambdaPath:   args.RuntimeLambdaPath,
		FunctionsLambdaPath: args.FunctionsLambdaPath,
		Env:                 args.Env,
		Config:              args.Config,
		Schema:              args.Schema,
	})

	// Add program to workspace options
	wsOptions := append(args.PulumiConfig.WorkspaceOptions, auto.Program(runFunc))

	ws, err := auto.NewLocalWorkspace(ctx, wsOptions...)
	if err != nil {
		log(ctx, "%s error creating workspace: %s", IconCross, err.Error())
		return nil, err
	}

	t := NewTiming()

	s, err := auto.UpsertStack(ctx, args.PulumiConfig.StackName, ws)
	if err != nil {
		log(ctx, "%s error selecting stack: %s", IconCross, err.Error())
		return nil, err
	}

	log(ctx, "%s Selected stack %s %s", IconTick, orange("%s", args.PulumiConfig.StackName), t.Since())

	// This may not be needed as we've already set the region on the workspace env vars but no harm in setting it here too
	err = s.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: args.Config.Deploy.Region})
	if err != nil {
		log(ctx, "%s error setting config aws:region on stack: %s", IconCross, err.Error())
		return nil, err
	}

	// Latest release can be found here - https://github.com/pulumi/pulumi-aws/releases
	awsPluginVersion := "v6.63.0"

	err = s.Workspace().InstallPlugin(ctx, "aws", awsPluginVersion)
	if err != nil {
		log(ctx, "%s error installing AWS Pulumi plugin:%s", IconCross, err.Error())
		return nil, err
	}

	log(ctx, "%s AWS plugin %s installed %s", IconTick, orange("%s", awsPluginVersion), t.Since())

	_, err = s.Refresh(ctx)
	if err != nil {
		log(ctx, "%s error refreshing stack: %s", IconCross, err.Error())
		return nil, err
	}

	log(ctx, "%s Stack refreshed %s", IconTick, t.Since())
	return &s, nil
}

const (
	StackOutputApiURL               = "apiUrl"
	StackOutputDatabaseEndpoint     = "databaseEndpoint"
	StackOutputDatabaseDbName       = "databaseDbName"
	StackOutputDatabaseSecretArn    = "databaseSecretArn"
	StackOutputApiLambdaName        = "apiLambdaName"
	StackOutputSubscriberLambdaName = "subscriberLambdaName"
	StackOutputJobsLambdaName       = "jobsLambdaName"
	StackOutputFunctionsLambdaName  = "functionsLambdaName"
)

type StackOutputs struct {
	ApiURL               string
	DatabaseEndpoint     string
	DatabaseDbName       string
	DatabaseSecretArn    string
	ApiLambdaName        string
	SubscriberLambdaName string
	JobsLambdaName       string
	FunctionsLambdaName  string
}

type GetStackOutputsArgs struct {
	Config       *config.ProjectConfig
	PulumiConfig *PulumiConfig
}

func getStackOutputs(ctx context.Context, args *GetStackOutputsArgs) (*StackOutputs, error) {
	ws, err := auto.NewLocalWorkspace(ctx, args.PulumiConfig.WorkspaceOptions...)
	if err != nil {
		log(ctx, "%s error creating Pulumi workspace: %s", IconCross, gray(err.Error()))
		return nil, err
	}

	stack, err := auto.SelectStack(ctx, args.PulumiConfig.StackName, ws)
	if auto.IsSelectStack404Error(err) {
		return nil, nil
	}
	if err != nil {
		log(ctx, "%s error selecting stack: %s", IconCross, gray(err.Error()))
		return nil, err
	}

	outputs, err := stack.Outputs(ctx)
	if err != nil {
		log(ctx, "%s error getting stack outputs: %s", IconCross, gray(err.Error()))
		return nil, err
	}

	return parseStackOutputs(outputs), nil
}

func parseStackOutputs(outputs auto.OutputMap) *StackOutputs {
	result := &StackOutputs{}
	for key, output := range outputs {
		v := output.Value.(string)
		switch key {
		case StackOutputApiURL:
			result.ApiURL = v
		case StackOutputDatabaseDbName:
			result.DatabaseDbName = v
		case StackOutputDatabaseEndpoint:
			result.DatabaseEndpoint = v
		case StackOutputDatabaseSecretArn:
			result.DatabaseSecretArn = v
		case StackOutputApiLambdaName:
			result.ApiLambdaName = v
		case StackOutputSubscriberLambdaName:
			result.SubscriberLambdaName = v
		case StackOutputJobsLambdaName:
			result.JobsLambdaName = v
		case StackOutputFunctionsLambdaName:
			result.FunctionsLambdaName = v
		}
	}

	return result
}
