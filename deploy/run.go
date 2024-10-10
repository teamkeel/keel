package deploy

import (
	"context"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/events"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/teamkeel/keel/schema"
)

const (
	UpAction      = "up"
	DestroyAction = "destroy"
)

type RunArgs struct {
	Action           string
	ProjectRoot      string
	Env              string
	RuntimeBinaryURL string
	Events           chan Output
}

func Run(ctx context.Context, args *RunArgs) error {
	defer func() {
		close(args.Events)
	}()

	args.Events <- Output{
		Heading: "Keel Config",
	}

	keelConfig, err := loadKeelConfig(args)
	if keelConfig == nil || err != nil {
		return err
	}

	args.Events <- Output{
		Heading: "Schema",
	}
	builder := schema.Builder{}
	protoSchema, err := builder.MakeFromDirectory(args.ProjectRoot)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Schema is not valid - run `keel validate` to see all validation errors",
			Error:   err,
		}
		return err
	}
	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "Schema is valid",
	}

	args.Events <- Output{
		Heading: "Build",
	}
	buildLambdasResult, err := buildLambdas(ctx, &BuildLambdasArgs{
		Config:           keelConfig,
		Schema:           protoSchema,
		RuntimeBinaryURL: args.RuntimeBinaryURL,
		ProjectRoot:      args.ProjectRoot,
		Events:           args.Events,
	})
	if err != nil {
		return err
	}

	args.Events <- Output{
		Heading: "AWS Config",
	}
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(keelConfig.Deploy.Region))
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error loading default AWS config",
			Error:   err,
		}
		return err
	}

	awsIdentity, err := getAwsIdentity(ctx, cfg, args.Events)
	if err != nil {
		return err
	}

	err = createPrivateKeySecret(ctx, &CreatePrivateKeySecretArgs{
		AwsConfig:   cfg,
		ProjectName: keelConfig.Deploy.ProjectName,
		Env:         args.Env,
		Events:      args.Events,
	})
	if err != nil {
		return err
	}

	args.Events <- Output{
		Heading: "Pulumi",
	}
	pulumiCfg := setupPulumi(ctx, &SetupPulumiArgs{
		AwsConfig: cfg,
		Config:    keelConfig,
		Env:       args.Env,
		Events:    args.Events,
	})
	if pulumiCfg == nil {
		return nil
	}
	stack, err := selectStack(ctx, &SelectStackArgs{
		Env:             args.Env,
		Schema:          protoSchema,
		Config:          keelConfig,
		AwsConfig:       cfg,
		AccountID:       awsIdentity.AccountID,
		RuntimeLambda:   buildLambdasResult.RuntimeLambda,
		FunctionsLambda: buildLambdasResult.FunctionsLambda,
		PulumiConfig:    pulumiCfg,
		Events:          args.Events,
	})
	if err != nil {
		return err
	}

	if args.Action == UpAction {
		args.Events <- Output{
			Heading: "Database Migrations Check",
		}
		err = runMigrations(ctx, &RunMigrationsArgs{
			AwsConfig: cfg,
			Stack:     stack,
			Schema:    protoSchema,
			DryRun:    true,
			Events:    args.Events,
		})
		if err != nil {
			return err
		}
	}

	args.Events <- Output{
		Heading: "Deploying",
	}

	eventsChannel := make(chan events.EngineEvent, 0)
	go func() {
		pending := map[string]bool{}
		for event := range eventsChannel {
			switch {
			case event.ResourcePreEvent != nil:
				if event.ResourcePreEvent.Metadata.Op != apitype.OpSame {
					pending[event.ResourcePreEvent.Metadata.URN] = true
					args.Events <- Output{
						Icon:    OutputIconPipe,
						Message: fmt.Sprintf("%s (%s)", cleanURN(event.ResourcePreEvent.Metadata.URN), event.ResourcePreEvent.Metadata.Op),
					}
				}
			case event.ResOutputsEvent != nil:
				if _, ok := pending[event.ResOutputsEvent.Metadata.URN]; ok {
					args.Events <- Output{
						Icon:    OutputIconPipe,
						Message: fmt.Sprintf("%s (done)", cleanURN(event.ResOutputsEvent.Metadata.URN)),
					}
				}
			case event.SummaryEvent != nil:
				args.Events <- Output{
					Icon:    OutputIconTick,
					Message: fmt.Sprintf("Duration %d seconds", event.SummaryEvent.DurationSeconds),
				}
			default:
				// for debugging
				// b, _ := json.Marshal(event)
				// fmt.Println("other event", string(b))
			}
		}
	}()

	if args.Action == DestroyAction {
		args.Events <- Output{
			Icon:    OutputIconPipe,
			Message: "Destroying stack",
		}

		// destroy our stack and exit early
		_, err := stack.Destroy(ctx, optdestroy.EventStreams(eventsChannel))
		if err != nil {
			args.Events <- Output{
				Icon:    OutputIconCross,
				Message: "Error destroying stack",
				Error:   err,
			}
			return err
		}

		args.Events <- Output{
			Icon:    OutputIconTick,
			Message: "Stack destroyed",
		}
		return nil
	}

	args.Events <- Output{
		Icon:    OutputIconPipe,
		Message: "Deploying stack",
	}
	res, err := stack.Up(ctx, optup.EventStreams(eventsChannel))
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error deploying stack",
			Error:   err,
		}
		return err
	}

	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "Stack deployed",
	}

	args.Events <- Output{
		Heading: "Database Migrations",
	}
	err = runMigrations(ctx, &RunMigrationsArgs{
		AwsConfig: cfg,
		Stack:     stack,
		Schema:    protoSchema,
		DryRun:    false,
		Events:    args.Events,
	})
	if err != nil {
		return err
	}

	// get the URL from the stack outputs
	url, ok := res.Outputs["apiUrl"].Value.(string)
	if !ok {
		err := fmt.Errorf("Error getting apiUrl from stack outputs")
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error getting apiUrl from stack outputs",
			Error:   err,
		}
		return err
	}

	args.Events <- Output{
		Heading: "Summary",
	}
	args.Events <- Output{
		Icon:    OutputIconPipe,
		Message: fmt.Sprintf("API URL: %s", url),
	}
	return nil
}

func cleanURN(s string) string {
	parts := strings.Split(s, "::")
	if len(parts) < 3 {
		return s
	}
	return strings.Join(parts[2:], "::")
}
