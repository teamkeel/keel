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

func selectStack(ctx context.Context, args *SelectStackArgs) (auto.Stack, error) {
	runFunc := createProgram(&NewProgramArgs{
		AwsConfig:           args.AwsConfig,
		AwsAccountID:        args.AccountID,
		RuntimeLambdaPath:   args.RuntimeLambdaPath,
		FunctionsLambdaPath: args.FunctionsLambdaPath,
		Env:                 args.Env,
		Config:              args.Config,
		Schema:              args.Schema,
	})

	t := NewTiming()
	s, err := auto.UpsertStackInlineSource(
		ctx,
		args.PulumiConfig.StackName,
		args.Config.Deploy.ProjectName,
		runFunc,
		args.PulumiConfig.WorkspaceOptions...)
	if err != nil {
		log("%s error selecting stack: %s", IconCross, err.Error())
		return s, err
	}

	log("%s Selected stack %s %s", IconTick, orange(args.PulumiConfig.StackName), t.Since())

	awsPluginVersion := "v6.63.0"

	// for inline source programs, we must manage plugins ourselves
	err = s.Workspace().InstallPlugin(ctx, "aws", awsPluginVersion)
	if err != nil {
		log("%s error installing AWS Pulumi plugin:%s", IconCross, err.Error())
		return s, err
	}

	log("%s AWS plugin %s installed %s", IconTick, orange(awsPluginVersion), t.Since())

	// set stack configuration specifying the AWS region to deploy
	err = s.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: args.Config.Deploy.Region})
	if err != nil {
		log("%s error setting aws:region on Pulumi stack: %s", IconCross, err.Error())
		return s, err
	}

	_, err = s.Refresh(ctx)
	if err != nil {
		log("%s error refreshing stack: %s", IconCross, err.Error())
		return s, err
	}

	log("%s Stack refreshed %s", IconTick, t.Since())
	return s, nil
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

func getStackOutputs(ctx context.Context, pulumiConfig *PulumiConfig) (*StackOutputs, error) {
	ws, err := auto.NewLocalWorkspace(ctx, pulumiConfig.WorkspaceOptions...)
	if err != nil {
		return nil, err
	}

	stack, err := auto.SelectStack(ctx, pulumiConfig.StackName, ws)
	if auto.IsSelectStack404Error(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	outputs, err := stack.Outputs(ctx)
	if err != nil {
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
