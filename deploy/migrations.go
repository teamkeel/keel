package deploy

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
)

type RunMigrationsArgs struct {
	AwsConfig aws.Config
	Stack     auto.Stack
	Schema    *proto.Schema
	DryRun    bool
	Events    chan Output
}

func runMigrations(ctx context.Context, args *RunMigrationsArgs) error {
	// TODO: support external database by fetching DATABASE_URL secret from SSM
	stackOutputs, err := args.Stack.Outputs(ctx)
	if err != nil {
		args.Events <- Output{
			Message: "Error getting stack outputs",
			Error:   err,
		}
		return err
	}

	outputs := parseStackOutputs(stackOutputs)

	if outputs.DatabaseSecretArn == "" {
		args.Events <- Output{
			Message: "No database secret ARN in stack outputs - skipping migrations check",
		}
		return nil
	}

	databaseURL, err := GetRDSConnection(ctx, &GetRDSConnectionArgs{
		Cfg:       args.AwsConfig,
		Endpoint:  outputs.DatabaseEndpoint,
		DbName:    outputs.DatabaseDbName,
		SecretArn: outputs.DatabaseSecretArn,
	})
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error getting RDS connection string",
			Error:   err,
		}
		return err
	}

	dbConn, err := db.New(ctx, databaseURL)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error connecting to database",
			Error:   err,
		}
		return err
	}

	m, err := migrations.New(ctx, args.Schema, dbConn)
	if err != nil {
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: "Error generating database migrations",
			Error:   err,
		}
		return err
	}

	err = m.Apply(ctx, args.DryRun)
	if err != nil {
		message := "Error applying database migrations"
		if args.DryRun {
			message = "Database migrations can not be applied"
		}
		if strings.Contains(err.Error(), "contains null values") {
			message = fmt.Sprintf(`%s.

The most likely cause of this is that you have added a non-null field to a model for which there are already rows in the database.
To fix this either make the new field optional (by using '?') or provide a default value using @default.`, message)
		}
		args.Events <- Output{
			Icon:    OutputIconCross,
			Message: message,
			Error:   err,
		}
		return err
	}

	if args.DryRun {
		args.Events <- Output{
			Icon:    OutputIconTick,
			Message: "Schema migrations can be applied",
		}
		if len(m.Changes) > 0 {
			for _, ch := range m.Changes {
				message := ch.Model
				if ch.Field != "" {
					message = fmt.Sprintf("%s.%s", message, ch.Field)
				}
				message = fmt.Sprintf("%s (%s)", message, ch.Type)
				args.Events <- Output{
					Icon:    OutputIconPipe,
					Message: message,
				}
			}
		}
		return nil
	}

	args.Events <- Output{
		Icon:    OutputIconTick,
		Message: "Database migrations run successfully",
	}
	return nil
}
