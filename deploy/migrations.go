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
}

func runMigrations(ctx context.Context, args *RunMigrationsArgs) error {
	t := NewTiming()

	// TODO: support external database by fetching DATABASE_URL secret from SSM
	stackOutputs, err := args.Stack.Outputs(ctx)
	if err != nil {
		log("%s error getting stack outputs: %s", IconCross, err.Error())
		return err
	}

	outputs := parseStackOutputs(stackOutputs)

	if outputs.DatabaseSecretArn == "" {
		log("%s no database secret ARN in stack outputs - skipping migrations check", IconPipe)
		return nil
	}

	databaseURL, err := GetRDSConnection(ctx, &GetRDSConnectionArgs{
		Cfg:       args.AwsConfig,
		Endpoint:  outputs.DatabaseEndpoint,
		DbName:    outputs.DatabaseDbName,
		SecretArn: outputs.DatabaseSecretArn,
	})
	if err != nil {
		log("%s error getting RDS connection string: %s", IconCross, err.Error())
		return err
	}

	dbConn, err := db.New(ctx, databaseURL)
	if err != nil {
		log("%s error  connecting to the database: %s", IconCross, err.Error())
		return err
	}

	m, err := migrations.New(ctx, args.Schema, dbConn)
	if err != nil {
		log("%s error generating database migrations: %s", IconCross, err.Error())
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
		log("%s %s: %s", IconCross, message, err.Error())
		return err
	}

	if args.DryRun {
		switch {
		case len(m.Changes) == 0:
			log("%s No database schema changes required %s", IconTick, t.Since())
		default:
			log("%s The following database schema changes will be applied %s", IconTick, t.Since())
			for _, ch := range m.Changes {
				message := ch.Model
				if ch.Field != "" {
					message = fmt.Sprintf("%s.%s", message, ch.Field)
				}

				action := ""
				switch ch.Type {
				case migrations.ChangeTypeAdded:
					action = green("(added)")
				case migrations.ChangeTypeRemoved:
					action = red("(removed)")
				case migrations.ChangeTypeModified:
					action = gray("(modified)")
				}

				log("    - %s %s", orange(message), action)
			}
		}
		return nil
	}

	switch {
	case len(m.Changes) == 0:
		log("%s Databaase schema is up-to-date %s", IconTick, t.Since())
	default:
		log("%s %s database schema changes applied %s", IconTick, orange("%d", len(m.Changes)), t.Since())
	}

	return nil
}
