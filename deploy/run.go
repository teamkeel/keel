package deploy

import (
	"context"
	"errors"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/events"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

const (
	UpAction     = "up"
	RemoveAction = "remove"
)

type RunArgs struct {
	Action        string
	ProjectRoot   string
	Env           string
	RuntimeBinary string
}

func Run(ctx context.Context, args *RunArgs) error {
	buildLambdasResult, err := Build(ctx, &BuildArgs{
		ProjectRoot:   args.ProjectRoot,
		Env:           args.Env,
		RuntimeBinary: args.RuntimeBinary,
	})
	if err != nil {
		return err
	}

	protoSchema := buildLambdasResult.Schema
	projectConfig := buildLambdasResult.Config

	deploy := projectConfig.Deploy
	if deploy == nil {
		err = fmt.Errorf("missing 'deploy' section in Keel config file")
		log("%s %s", IconCross, err.Error())
		return err
	}

	if args.Action == UpAction {
		heading("Deploy")
	} else {
		heading("Remove")
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(projectConfig.Deploy.Region))
	if err != nil {
		log("%s error loading AWS config", IconCross, err.Error())
		return err
	}

	awsIdentity, err := getAwsIdentity(ctx, cfg)
	if err != nil {
		return err
	}

	log("%s AWS account %s", IconTick, orange(awsIdentity.AccountID))

	if args.Action == UpAction {
		err = createPrivateKeySecret(ctx, &CreatePrivateKeySecretArgs{
			AwsConfig:   cfg,
			ProjectName: projectConfig.Deploy.ProjectName,
			Env:         args.Env,
		})
		if err != nil {
			return err
		}
	}

	pulumiCfg, err := setupPulumi(ctx, &SetupPulumiArgs{
		AwsConfig: cfg,
		Config:    projectConfig,
		Env:       args.Env,
	})
	if err != nil {
		return err
	}

	stack, err := selectStack(ctx, &SelectStackArgs{
		Env:                 args.Env,
		Schema:              protoSchema,
		Config:              projectConfig,
		AwsConfig:           cfg,
		AccountID:           awsIdentity.AccountID,
		RuntimeLambdaPath:   buildLambdasResult.RuntimePath,
		FunctionsLambdaPath: buildLambdasResult.FunctionsPath,
		PulumiConfig:        pulumiCfg,
	})
	if err != nil {
		return err
	}

	if args.Action == UpAction {
		err = runMigrations(ctx, &RunMigrationsArgs{
			AwsConfig: cfg,
			Stack:     stack,
			Env:       args.Env,
			Schema:    protoSchema,
			Config:    projectConfig,
			DryRun:    true,
		})
		if err != nil {
			return err
		}
	}

	eventsChannel := make(chan events.EngineEvent)
	go func() {
		pending := map[string]*Timing{}
		for event := range eventsChannel {
			switch {
			case event.ResourcePreEvent != nil:
				urn := event.ResourcePreEvent.Metadata.URN
				// We don't bother showing anything if a resource hasn't changed
				if event.ResourcePreEvent.Metadata.Op != apitype.OpSame {
					pending[urn] = NewTiming()
					urn := cleanURN(urn)
					log("%s %s - %s", IconPipe, gray(urn), orange("%s", event.ResourcePreEvent.Metadata.Op))
				}
			case event.ResOutputsEvent != nil:
				urn := event.ResOutputsEvent.Metadata.URN
				if t, ok := pending[urn]; ok {
					urn := cleanURN(urn)
					log("%s %s - %s %s", IconPipe, gray(urn), green("done"), t.Since())
				}
			default:
				// for debugging other events...
				// b, _ := json.Marshal(event)
				// log("Pulumi event: %s", string(b))
			}
		}
	}()

	t := NewTiming()

	if args.Action == RemoveAction {
		log("%s Removing resources...", IconPipe)
		_, err := stack.Destroy(ctx, optdestroy.EventStreams(eventsChannel))
		if err != nil {
			log("%s Error removing stack: %s", IconCross, err.Error())
			return err
		}

		log("%s Stack removed %s", IconTick, t.Since())
		return nil
	}

	log("%s Updating resources...", IconPipe)
	res, err := stack.Up(ctx, optup.EventStreams(eventsChannel))
	if err != nil {
		log("%s error deploying stack: %s", IconCross, err.Error())
		return err
	}

	log("%s App successfully deployed %s", IconTick, t.Since())

	err = runMigrations(ctx, &RunMigrationsArgs{
		AwsConfig: cfg,
		Stack:     stack,
		Env:       args.Env,
		Schema:    protoSchema,
		Config:    projectConfig,
		DryRun:    false,
	})
	if err != nil {
		return err
	}

	// Get the URL from the stack outputs
	url, ok := res.Outputs["apiUrl"].Value.(string)
	if !ok {
		err := fmt.Errorf("error getting API url from stack outputs")
		log("%s %s", IconCross, err.Error())
		return err
	}

	log("%s API URL: %s", IconTick, orange(url))
	return nil
}

func cleanURN(s string) string {
	parts := strings.Split(s, "::")
	if len(parts) < 3 {
		return s
	}
	return strings.Join(parts[2:], "::")
}

func isSmithyAPIError(err error, code string) bool {
	var ae smithy.APIError
	return errors.As(err, &ae) && ae.ErrorCode() == code
}
