package deploy

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
)

type SelectStackArgs struct {
	AwsConfig       aws.Config
	PulumiConfig    *PulumiConfig
	Env             string
	Schema          *proto.Schema
	Config          *config.ProjectConfig
	AccountID       string
	RuntimeLambda   pulumi.Archive
	FunctionsLambda pulumi.Archive
	Events          chan Output
}

func selectStack(ctx context.Context, args *SelectStackArgs) (auto.Stack, error) {
	runFunc := createProgram(&NewProgramArgs{
		AwsConfig:       args.AwsConfig,
		AwsAccountID:    args.AccountID,
		RuntimeLambda:   args.RuntimeLambda,
		FunctionsLambda: args.FunctionsLambda,
		Env:             args.Env,
		Config:          args.Config,
		Schema:          args.Schema,
	})
	s, err := auto.UpsertStackInlineSource(
		ctx,
		args.PulumiConfig.StackName,
		args.Config.Deploy.ProjectName,
		runFunc,
		args.PulumiConfig.WorkspaceOptions...)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Failed to setup wokspace",
			Error:   err,
		}
		return s, err
	}

	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: fmt.Sprintf("Selected stack: %s", args.PulumiConfig.StackName),
	}

	// for inline source programs, we must manage plugins ourselves
	err = s.Workspace().InstallPlugin(ctx, "aws", "v6.55.0")
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Failed to install AWS plugin",
			Error:   err,
		}
		return s, err
	}

	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "AWS plugin v6.55.0 installed",
	}

	// set stack configuration specifying the AWS region to deploy
	err = s.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: args.Config.Deploy.Region})
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error setting aws:region on stack",
			Error:   err,
		}
		return s, err
	}

	_, err = s.Refresh(ctx)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error refreshing stack",
			Error:   err,
		}
		return s, err
	}

	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "Stack refreshed",
	}
	return s, nil
}

const (
	StackOutputApiURL               = "apiUrl"
	StackOutputDatabaseEndpoint     = "databaseEndpoint"
	StackOutputDatabaseDbName       = "databaseDbName"
	StackOutputDatabaseSecretArn    = "databaseSecretArn"
	StackOutputApiLambdaName        = "apiLambdaName"
	StackOutputSubscriberLambdaName = "subscriberLambdaName"
	StackOutputFunctionsLambdaName  = "functionsLambdaName"
)

type StackOutputs struct {
	ApiURL               string
	DatabaseEndpoint     string
	DatabaseDbName       string
	DatabaseSecretArn    string
	ApiLambdaName        string
	SubscriberLambdaName string
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
		case StackOutputFunctionsLambdaName:
			result.FunctionsLambdaName = v
		case StackOutputSubscriberLambdaName:
			result.SubscriberLambdaName = v
		}
	}

	return result
}
