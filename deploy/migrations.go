package deploy

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/deploy/lambdas/runtime"
	"github.com/teamkeel/keel/migrations"
	"github.com/teamkeel/keel/proto"
)

type RunMigrationsArgs struct {
	AwsConfig aws.Config
	Stack     *auto.Stack
	Env       string
	Schema    *proto.Schema
	Config    *config.ProjectConfig
	DryRun    bool
}

func runMigrations(ctx context.Context, args *RunMigrationsArgs) error {
	t := NewTiming()

	var databaseURL string

	deployConf := args.Config.Deploy
	dbConf := deployConf.Database

	if dbConf != nil && dbConf.Provider == "external" {
		paramName := runtime.SsmParameterName(deployConf.ProjectName, args.Env, "DATABASE_URL")
		res, err := ssm.NewFromConfig(args.AwsConfig).GetParameter(ctx, &ssm.GetParameterInput{
			Name:           aws.String(paramName),
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			log(ctx, "%s Error fetching %s secret from SSM: %s", IconCross, orange("DATABASE_URL"), gray("%s", err.Error()))
			return err
		}

		databaseURL = *res.Parameter.Value
	} else {
		stackOutputs, err := args.Stack.Outputs(ctx)
		if err != nil {
			log(ctx, "%s error getting stack outputs: %s", IconCross, err.Error())
			return err
		}

		outputs := parseStackOutputs(stackOutputs)

		if outputs.DatabaseSecretArn == "" {
			log(ctx, "%s no database secret ARN in stack outputs - skipping migrations check", IconPipe)
			return nil
		}

		databaseURL, err = runtime.GetRDSConnection(ctx, &runtime.GetRDSConnectionArgs{
			Cfg:       args.AwsConfig,
			Endpoint:  outputs.DatabaseEndpoint,
			DbName:    outputs.DatabaseDbName,
			SecretArn: outputs.DatabaseSecretArn,
		})
		if err != nil {
			log(ctx, "%s Error getting RDS connection string: %s", IconCross, gray("%s", err.Error()))
			return err
		}
	}

	dbConn, err := db.New(ctx, databaseURL)
	if err != nil {
		log(ctx, "%s Error connecting to the database: %s", IconCross, gray(err.Error()))
		return err
	}

	m, err := migrations.New(ctx, args.Schema, dbConn)
	if err != nil {
		log(ctx, "%s Error generating database migrations: %s", IconCross, gray(err.Error()))
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
		log(ctx, "%s %s: %s", IconCross, message, err.Error())
		return err
	}

	if args.DryRun {
		switch {
		case len(m.Changes) == 0:
			log(ctx, "%s No database schema changes required %s", IconTick, t.Since())
		default:
			log(ctx, "%s The following database schema changes will be applied %s", IconTick, t.Since())
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

				log(ctx, "    - %s %s", orange(message), action)
			}
		}
		return nil
	}

	switch {
	case len(m.Changes) == 0:
		log(ctx, "%s Database schema is up-to-date %s", IconTick, t.Since())
	default:
		log(ctx, "%s %s database schema changes applied %s", IconTick, orange("%d", len(m.Changes)), t.Since())
	}

	return nil
}
