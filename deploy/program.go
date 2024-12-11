package deploy

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	pulumiaws "github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/rds"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/scheduler"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/sqs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/deploy/lambdas/runtime"
	"github.com/teamkeel/keel/proto"
)

// https://github.com/open-telemetry/opentelemetry-lambda/releases/tag/layer-collector%2F0.12.0
const otelCollectorLayer = "arn:aws:lambda:%s:184161586896:layer:opentelemetry-collector-amd64-0_12_0:1"

type NewProgramArgs struct {
	AwsConfig           aws.Config
	AwsAccountID        string
	RuntimeLambdaPath   string
	FunctionsLambdaPath string
	Env                 string
	Config              *config.ProjectConfig
	Schema              *proto.Schema
}

func createProgram(args *NewProgramArgs) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		deployCfg := args.Config.Deploy

		projectName := deployCfg.ProjectName
		region := deployCfg.Region

		baseTags := pulumi.StringMap{
			"Project": pulumi.String(projectName),
			"Env":     pulumi.String(args.Env),
		}

		externalDB := deployCfg.Database != nil && deployCfg.Database.Provider == "external"
		var rds *CreateRDSResourcesResult

		if !externalDB {
			var err error
			rds, err = createRDSResources(ctx, &CreateRDSResourcesArgs{
				BaseTags: baseTags,
				Env:      args.Env,
				Config:   args.Config,
			})
			if err != nil {
				return err
			}
		}

		filesBucket, err := s3.NewBucket(ctx, "file-storage", &s3.BucketArgs{
			BucketPrefix: pulumi.StringPtr(fmt.Sprintf("%s--%s-", projectName, args.Env)),
			Tags:         baseTags,
		}, pulumi.RetainOnDelete(true))
		if err != nil {
			return err
		}

		queue, err := sqs.NewQueue(ctx, "events", &sqs.QueueArgs{
			Tags: baseTags,
		})
		if err != nil {
			return err
		}

		functionsRole, err := createLambdaRole(ctx, "functions", iam.GetPolicyDocumentStatementArray{
			iam.GetPolicyDocumentStatementInput(iam.GetPolicyDocumentStatementArgs{
				Actions: pulumi.ToStringArray([]string{
					"s3:GetObject",
					"s3:PutObject",
					"s3:DeleteObject",
				}),
				Resources: pulumi.ToStringArrayOutput(
					[]pulumi.StringOutput{
						filesBucket.Arn.ApplyT(func(v string) string {
							return v + "/*"
						}).(pulumi.StringOutput),
					},
				),
			}),
		}, baseTags)
		if err != nil {
			return err
		}

		tracingEnabled := deployCfg.Telemetry != nil && deployCfg.Telemetry.Collector != ""

		// OTEL config
		var otelLayer pulumi.StringArray
		if tracingEnabled {
			arn := fmt.Sprintf(otelCollectorLayer, region)
			otelLayer = pulumi.ToStringArray([]string{arn})
		}

		functionEnvVars := pulumi.StringMap{
			"KEEL_PROJECT_NAME":      pulumi.String(projectName),
			"KEEL_ENV":               pulumi.String(args.Env),
			"KEEL_FILES_BUCKET_NAME": filesBucket.Bucket,
			// The actual connection string is passed from the runtime to functions
			// via a secret
			"KEEL_DB_CONN_TYPE": pulumi.String("pg"),
			"KEEL_DB_CERT":      pulumi.String("/var/task/rds.pem"),
			"NODE_OPTIONS":      pulumi.String("--enable-source-maps"),
		}

		// OTEL config
		if tracingEnabled {
			functionEnvVars = extendStringMap(functionEnvVars, pulumi.StringMap{
				"OPENTELEMETRY_COLLECTOR_CONFIG_URI": pulumi.String("/var/task/collector.yaml"),
				"KEEL_TRACING_ENABLED":               pulumi.String("true"),
			})
		}

		// Add env vars from config
		for _, env := range args.Config.Environment {
			functionEnvVars[env.Name] = pulumi.String(env.Value)
		}

		functions, err := lambda.NewFunction(ctx, "functions", &lambda.FunctionArgs{
			Code:       pulumi.NewFileArchive(args.FunctionsLambdaPath),
			Runtime:    lambda.RuntimeNodeJS20dX,
			MemorySize: pulumi.IntPtr(2048),
			Role:       functionsRole.Arn,
			Handler:    pulumi.String("main-bundled.handler"),
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: functionEnvVars,
			},
			Layers: otelLayer,
			LoggingConfig: lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("JSON"),
			},
			Tags: baseTags,
		})
		if err != nil {
			return fmt.Errorf("error creating runtime lambda: %v", err)
		}

		runtimePolicyStatements := iam.GetPolicyDocumentStatementArray{
			iam.GetPolicyDocumentStatementInput(iam.GetPolicyDocumentStatementArgs{
				Actions: pulumi.ToStringArray([]string{
					"ssm:GetParameter",
					"ssm:GetParameters",
				}),
				Resources: pulumi.ToStringArray([]string{
					fmt.Sprintf("arn:aws:ssm:%s:%s:parameter%s",
						region,
						args.AwsAccountID,
						runtime.SsmParameterName(projectName, args.Env, "*")),
				}),
			}),
			iam.GetPolicyDocumentStatementInput(iam.GetPolicyDocumentStatementArgs{
				Actions: pulumi.ToStringArray([]string{
					"s3:GetObject",
					"s3:PutObject",
					"s3:DeleteObject",
				}),
				Resources: pulumi.ToStringArrayOutput(
					[]pulumi.StringOutput{
						filesBucket.Arn.ApplyT(func(v string) string {
							return v + "/*"
						}).(pulumi.StringOutput),
					},
				),
			}),
			iam.GetPolicyDocumentStatementInput(iam.GetPolicyDocumentStatementArgs{
				Actions: pulumi.ToStringArray([]string{
					"lambda:InvokeFunction",
				}),
				Resources: pulumi.ToStringArrayOutput(
					[]pulumi.StringOutput{
						functions.Arn,
					},
				),
			}),
			iam.GetPolicyDocumentStatementInput(iam.GetPolicyDocumentStatementArgs{
				Actions: pulumi.ToStringArray([]string{
					"sqs:SendMessage",
					"sqs:GetQueueUrl",
					"sqs:DeleteMessage",
					"sqs:ReceiveMessage",
					"sqs:GetQueueAttributes",
				}),
				Resources: pulumi.ToStringArrayOutput(
					[]pulumi.StringOutput{
						queue.Arn,
					},
				),
			}),
		}

		if rds != nil {
			runtimePolicyStatements = append(runtimePolicyStatements, iam.GetPolicyDocumentStatementInput(iam.GetPolicyDocumentStatementArgs{
				Actions: pulumi.ToStringArray([]string{
					"secretsmanager:GetSecretValue",
				}),
				Resources: pulumi.ToStringArrayOutput([]pulumi.StringOutput{
					rds.SecretARN.Elem(),
				}),
			}))
		}

		runtimeRole, err := createLambdaRole(ctx, "runtime", runtimePolicyStatements, baseTags)
		if err != nil {
			return err
		}

		jobsWebhookURL := ""
		if deployCfg.Jobs != nil {
			jobsWebhookURL = deployCfg.Jobs.WebhookURL
		}

		secretNames := lo.Map(args.Config.Secrets, func(s config.Secret, _ int) string {
			return s.Name
		})
		secretNames = append(secretNames, "KEEL_PRIVATE_KEY")

		baseRuntimeEnvVars := pulumi.StringMap{
			"KEEL_PROJECT_NAME":      pulumi.String(projectName),
			"KEEL_ENV":               pulumi.String(args.Env),
			"KEEL_SECRETS":           pulumi.String(strings.Join(secretNames, ":")),
			"KEEL_FILES_BUCKET_NAME": filesBucket.Bucket,
			"KEEL_FUNCTIONS_ARN":     functions.Arn,
			"KEEL_QUEUE_URL":         queue.Url,
			"KEEL_JOBS_WEBHOOK_URL":  pulumi.String(jobsWebhookURL),
		}

		// RDS env vars
		if rds != nil {
			baseRuntimeEnvVars = extendStringMap(baseRuntimeEnvVars, pulumi.StringMap{
				"KEEL_DATABASE_ENDPOINT":   rds.Instance.Endpoint,
				"KEEL_DATABASE_DB_NAME":    rds.Instance.DbName,
				"KEEL_DATABASE_SECRET_ARN": rds.SecretARN.Elem(),
			})
		}

		// OTEL config
		if tracingEnabled {
			baseRuntimeEnvVars["OPENTELEMETRY_COLLECTOR_CONFIG_URI"] = pulumi.String("/var/task/collector.yaml")
			baseRuntimeEnvVars["KEEL_TRACING_ENABLED"] = pulumi.String("true")
		}

		// Add env vars from config
		for _, env := range args.Config.Environment {
			baseRuntimeEnvVars[env.Name] = pulumi.String(env.Value)
		}

		api, err := lambda.NewFunction(ctx, "api", &lambda.FunctionArgs{
			Code:       pulumi.NewFileArchive(args.RuntimeLambdaPath),
			Runtime:    lambda.RuntimeCustomAL2023,
			MemorySize: pulumi.IntPtr(2048),
			Role:       runtimeRole.Arn,
			Handler:    pulumi.String("main"),
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: extendStringMap(baseRuntimeEnvVars, pulumi.StringMap{
					"KEEL_RUNTIME_MODE": pulumi.String(runtime.RuntimeModeApi),
				}),
			},
			Layers: otelLayer,
			LoggingConfig: lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("JSON"),
			},
			Tags: baseTags,
		})
		if err != nil {
			return fmt.Errorf("error creating api lambda: %v", err)
		}

		subscriber, err := lambda.NewFunction(ctx, "subscriber", &lambda.FunctionArgs{
			Code:       pulumi.NewFileArchive(args.RuntimeLambdaPath),
			Runtime:    lambda.RuntimeCustomAL2023,
			MemorySize: pulumi.IntPtr(2048),
			Role:       runtimeRole.Arn,
			Handler:    pulumi.String("main"),
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: extendStringMap(baseRuntimeEnvVars, pulumi.StringMap{
					"KEEL_RUNTIME_MODE": pulumi.String(runtime.RuntimeModeSubscriber),
				}),
			},
			Layers: otelLayer,
			LoggingConfig: lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("JSON"),
			},
			Tags: baseTags,
		})
		if err != nil {
			return fmt.Errorf("error creating subscriber lambda: %v", err)
		}

		_, err = lambda.NewEventSourceMapping(ctx, "subscriber-event-source-mapping", &lambda.EventSourceMappingArgs{
			EventSourceArn: queue.Arn,
			FunctionName:   subscriber.Arn,
			Tags:           baseTags,
		})
		if err != nil {
			return err
		}

		jobs, err := lambda.NewFunction(ctx, "jobs", &lambda.FunctionArgs{
			Code:       pulumi.NewFileArchive(args.RuntimeLambdaPath),
			Runtime:    lambda.RuntimeCustomAL2023,
			MemorySize: pulumi.IntPtr(2048),
			Role:       runtimeRole.Arn,
			Handler:    pulumi.String("main"),
			Layers:     otelLayer,
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: extendStringMap(baseRuntimeEnvVars, pulumi.StringMap{
					"KEEL_RUNTIME_MODE": pulumi.String(runtime.RuntimeModeJob),
				}),
			},
			LoggingConfig: lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("JSON"),
			},
			Tags: baseTags,
		})
		if err != nil {
			return fmt.Errorf("error creating jobs lambda: %v", err)
		}

		for _, job := range args.Schema.Jobs {
			if job.Schedule == nil {
				continue
			}

			expression := fmt.Sprintf("cron(%s)", strings.ReplaceAll(job.Schedule.Expression, "\"", ""))

			jobJson, err := json.Marshal(map[string]any{
				"name": job.Name,
			})
			if err != nil {
				return err
			}

			schedule, err := scheduler.NewSchedule(ctx, fmt.Sprintf("scheduled-job-%s", job.Name), &scheduler.ScheduleArgs{
				ScheduleExpression: pulumi.String(expression),
				FlexibleTimeWindow: scheduler.ScheduleFlexibleTimeWindowArgs{
					Mode: pulumi.String("OFF"),
				},
				Target: &scheduler.ScheduleTargetArgs{
					Arn:     jobs.Arn,
					RoleArn: runtimeRole.Arn,
					Input:   pulumi.StringPtr(string(jobJson)),
				},
				StartDate: pulumi.StringPtr(time.Now().UTC().Format(time.RFC3339)),
			})
			if err != nil {
				return err
			}

			_, err = lambda.NewPermission(ctx, fmt.Sprintf("scheduled-job-%s-permission", job.Name), &lambda.PermissionArgs{
				Action:    pulumi.String("lambda:InvokeFunction"),
				Function:  jobs.Name,
				Principal: pulumi.String("scheduler.amazonaws.com"),
				SourceArn: schedule.Arn,
			})
			if err != nil {
				return err
			}
		}

		apiURL, err := lambda.NewFunctionUrl(ctx, "api-url", &lambda.FunctionUrlArgs{
			AuthorizationType: pulumi.String("NONE"),
			Cors: &lambda.FunctionUrlCorsArgs{
				AllowCredentials: pulumi.BoolPtr(true),
				AllowHeaders:     pulumi.ToStringArray([]string{"*"}),
				AllowMethods:     pulumi.ToStringArray([]string{"*"}),
				AllowOrigins:     pulumi.ToStringArray([]string{"*"}),
				ExposeHeaders:    pulumi.ToStringArray([]string{"*"}),
			},
			FunctionName: api.Name,
		})
		if err != nil {
			return fmt.Errorf("error creating runtime lambda URL: %v", err)
		}

		ctx.Export(StackOutputApiURL, apiURL.FunctionUrl)
		ctx.Export(StackOutputApiLambdaName, api.Name)
		ctx.Export(StackOutputSubscriberLambdaName, subscriber.Name)
		ctx.Export(StackOutputFunctionsLambdaName, functions.Name)
		return nil
	}
}

type CreateRDSResourcesArgs struct {
	BaseTags pulumi.StringMap
	Env      string
	Config   *config.ProjectConfig
}

type CreateRDSResourcesResult struct {
	Instance  *rds.Instance
	SecretARN pulumi.StringPtrOutput
}

func createRDSResources(ctx *pulumi.Context, args *CreateRDSResourcesArgs) (*CreateRDSResourcesResult, error) {
	baseTags := args.BaseTags
	projectName := args.Config.Deploy.ProjectName

	vpc, err := ec2.NewVpc(ctx, "vpc", &ec2.VpcArgs{
		CidrBlock:          pulumi.String("10.0.0.0/16"),
		EnableDnsHostnames: pulumi.BoolPtr(true),
		Tags: extendStringMap(baseTags, pulumi.StringMap{
			"Name": pulumi.String(fmt.Sprintf("keel-%s-%s-vpc", projectName, args.Env)),
		}),
	})
	if err != nil {
		return nil, err
	}

	igw, err := ec2.NewInternetGateway(ctx, "internet-gateway", &ec2.InternetGatewayArgs{
		VpcId: vpc.ID(),
		Tags: extendStringMap(baseTags, pulumi.StringMap{
			"Name": pulumi.String(fmt.Sprintf("keel-%s-%s-internet-gateway", projectName, args.Env)),
		}),
	})
	if err != nil {
		return nil, err
	}

	azs := pulumiaws.GetAvailabilityZonesOutput(ctx, pulumiaws.GetAvailabilityZonesOutputArgs{
		State: pulumi.String("available"),
	})
	subnetIDs := azs.Names().ApplyT(func(names []string) (pulumi.StringArray, error) {
		result := pulumi.StringArray{}

		for i, zone := range names[:2] {
			subnet, err := ec2.NewSubnet(ctx, fmt.Sprintf("subnet-%d", i+1), &ec2.SubnetArgs{
				VpcId:               vpc.ID(),
				CidrBlock:           pulumi.String(fmt.Sprintf("10.0.%d.0/22", 8*i)),
				MapPublicIpOnLaunch: pulumi.Bool(true),
				AvailabilityZone:    pulumi.String(zone),
				Tags: extendStringMap(baseTags, pulumi.StringMap{
					"Name": pulumi.String(fmt.Sprintf("keel-%s-%s-subnet-%d", projectName, args.Env, i+1)),
				}),
			})
			if err != nil {
				return nil, err
			}

			routeTable, err := ec2.NewRouteTable(ctx, fmt.Sprintf("route-table-%d", i+1), &ec2.RouteTableArgs{
				VpcId: vpc.ID(),
				Routes: ec2.RouteTableRouteArray{
					&ec2.RouteTableRouteArgs{
						CidrBlock: pulumi.String("0.0.0.0/0"), // Route all traffic
						GatewayId: igw.ID(),                   // To the Internet Gateway
					},
				},
				Tags: extendStringMap(baseTags, pulumi.StringMap{
					"Name": pulumi.String(fmt.Sprintf("keel-%s-%s-route-table-%d", projectName, args.Env, i+1)),
				}),
			})
			if err != nil {
				return nil, err
			}

			_, err = ec2.NewRouteTableAssociation(ctx, fmt.Sprintf("route-table-association-%d", i+1), &ec2.RouteTableAssociationArgs{
				SubnetId:     subnet.ID(),
				RouteTableId: routeTable.ID(),
			})
			if err != nil {
				return nil, err
			}

			result = append(result, subnet.ID().ToStringOutput())
		}

		return result, nil
	}).(pulumi.StringArrayOutput)

	dbSubnetGroup, err := rds.NewSubnetGroup(ctx, "subnet-group", &rds.SubnetGroupArgs{
		SubnetIds: subnetIDs,
		Tags:      baseTags,
	})
	if err != nil {
		return nil, err
	}

	securityGroup, err := ec2.NewSecurityGroup(ctx, "db-security-group", &ec2.SecurityGroupArgs{
		VpcId: vpc.ID(),
		Ingress: ec2.SecurityGroupIngressArray{
			&ec2.SecurityGroupIngressArgs{
				Protocol:   pulumi.String("tcp"),
				FromPort:   pulumi.Int(5432),
				ToPort:     pulumi.Int(5432),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Egress: ec2.SecurityGroupEgressArray{
			&ec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
		Tags: baseTags,
	})
	if err != nil {
		return nil, err
	}

	// default to the cheapest instance available
	rdsInstanceType := "db.t4g.micro"
	rdsStorage := 20
	rdsMultiAz := false

	// update from config
	if args.Config.Deploy.Database != nil {
		db := args.Config.Deploy.Database
		if db.RDS != nil && db.RDS.Instance != nil {
			rdsInstanceType = *db.RDS.Instance
		}
		if db.RDS != nil && db.RDS.Storage != nil {
			rdsStorage = *db.RDS.Storage
		}
		if db.RDS != nil && db.RDS.MultiAZ != nil {
			rdsMultiAz = *db.RDS.MultiAZ
		}
	}

	// Create an RDS PostgreSQL instance
	dbInstance, err := rds.NewInstance(ctx, "rds", &rds.InstanceArgs{
		Engine:            pulumi.String("postgres"),
		InstanceClass:     pulumi.String(rdsInstanceType),
		AllocatedStorage:  pulumi.Int(rdsStorage),
		MultiAz:           pulumi.BoolPtr(rdsMultiAz),
		DbSubnetGroupName: dbSubnetGroup.Name,
		VpcSecurityGroupIds: pulumi.StringArray{
			securityGroup.ID(),
		},
		DbName:                   pulumi.String("keel"),
		Username:                 pulumi.String("keel"),
		ManageMasterUserPassword: pulumi.BoolPtr(true),
		SkipFinalSnapshot:        pulumi.Bool(true),
		PubliclyAccessible:       pulumi.Bool(true),
		Tags:                     baseTags,
	})
	if err != nil {
		return nil, err
	}

	dbSecretArn := dbInstance.MasterUserSecrets.Index(pulumi.Int(0)).SecretArn()

	ctx.Export(StackOutputDatabaseEndpoint, dbInstance.Endpoint)
	ctx.Export(StackOutputDatabaseDbName, dbInstance.DbName)
	ctx.Export(StackOutputDatabaseSecretArn, dbSecretArn)

	return &CreateRDSResourcesResult{
		Instance:  dbInstance,
		SecretARN: dbSecretArn,
	}, nil
}

func createLambdaRole(ctx *pulumi.Context, prefix string, statements iam.GetPolicyDocumentStatementArray, tags pulumi.StringMap) (*iam.Role, error) {
	role, err := iam.NewRole(ctx, fmt.Sprintf("%s-role", prefix), &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Action": "sts:AssumeRole",
					"Principal": {
						"Service": "lambda.amazonaws.com"
					},
					"Effect": "Allow",
					"Sid": ""
				},
				{
					"Action": "sts:AssumeRole",
					"Principal": {
						"Service": "scheduler.amazonaws.com"
					},
					"Effect": "Allow",
					"Sid": ""
				}
			]
		}`),
		Tags: tags,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating IAM role: %v", err)
	}

	lambdaPolicy, err := iam.NewPolicy(ctx, fmt.Sprintf("%s-policy", prefix), &iam.PolicyArgs{
		Policy: iam.GetPolicyDocumentOutput(ctx, iam.GetPolicyDocumentOutputArgs{
			Statements: statements,
		}).Json(),
		Tags: tags,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating role policy: %v", err)
	}

	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-managed-policy", prefix), &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: pulumi.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
	})
	if err != nil {
		return nil, fmt.Errorf("error attaching managed policy to role: %v", err)
	}

	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-policy-attachement", prefix), &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: lambdaPolicy.Arn,
	})
	if err != nil {
		return nil, fmt.Errorf("error attaching custom policy to role: %v", err)
	}

	return role, nil
}

func extendStringMap(a, b pulumi.StringMap) pulumi.StringMap {
	r := pulumi.StringMap{}
	for k, v := range a {
		r[k] = v
	}
	for k, v := range b {
		r[k] = v
	}
	return r
}
