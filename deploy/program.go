package deploy

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	pulumiaws "github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/lambda"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/rds"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/s3"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/sqs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
)

const (
	RuntimeModeApi        = "api"
	RuntimeModeSubscriber = "subscriber"
)

type NewProgramArgs struct {
	AwsConfig       aws.Config
	AwsAccountID    string
	RuntimeLambda   pulumi.Archive
	FunctionsLambda pulumi.Archive
	Env             string
	Config          *config.ProjectConfig
	Schema          *proto.Schema
}

func createProgram(args *NewProgramArgs) pulumi.RunFunc {
	return func(ctx *pulumi.Context) error {
		projectName := args.Config.Deploy.ProjectName
		region := args.Config.Deploy.Region

		baseTags := pulumi.StringMap{
			"Project": pulumi.String(args.Config.Deploy.ProjectName),
			"Env":     pulumi.String(args.Env),
		}

		vpc, err := ec2.NewVpc(ctx, "vpc", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.0.0.0/16"),
			EnableDnsHostnames: pulumi.BoolPtr(true),
			Tags: extendStringMap(baseTags, pulumi.StringMap{
				"Name": pulumi.String(fmt.Sprintf("keel-%s-%s-vpc", projectName, args.Env)),
			}),
		})
		if err != nil {
			return err
		}

		igw, err := ec2.NewInternetGateway(ctx, "internet-gateway", &ec2.InternetGatewayArgs{
			VpcId: vpc.ID(),
			Tags: extendStringMap(baseTags, pulumi.StringMap{
				"Name": pulumi.String(fmt.Sprintf("keel-%s-%s-internet-gateway", projectName, args.Env)),
			}),
		})
		if err != nil {
			return err
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
			return err
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
			return err
		}

		// Create an RDS PostgreSQL instance
		dbInstance, err := rds.NewInstance(ctx, "rds", &rds.InstanceArgs{
			Engine:            pulumi.String("postgres"),
			InstanceClass:     pulumi.String("db.t4g.micro"),
			AllocatedStorage:  pulumi.Int(20),
			DbSubnetGroupName: dbSubnetGroup.Name,
			VpcSecurityGroupIds: pulumi.StringArray{
				securityGroup.ID(),
			},
			MultiAz:                  pulumi.BoolPtr(false),
			DbName:                   pulumi.String("keel"),
			Username:                 pulumi.String("keel"),
			ManageMasterUserPassword: pulumi.BoolPtr(true),
			SkipFinalSnapshot:        pulumi.Bool(true),
			PubliclyAccessible:       pulumi.Bool(true),
			Tags:                     baseTags,
		})
		if err != nil {
			return err
		}

		dbSecretArn := dbInstance.MasterUserSecrets.Index(pulumi.Int(0)).SecretArn()

		ctx.Export(StackOutputDatabaseEndpoint, dbInstance.Endpoint)
		ctx.Export(StackOutputDatabaseDbName, dbInstance.DbName)
		ctx.Export(StackOutputDatabaseSecretArn, dbSecretArn)

		filesBucket, err := s3.NewBucket(ctx, "file-storage", &s3.BucketArgs{
			BucketPrefix: pulumi.StringPtr(fmt.Sprintf("%s--%s-", args.Config.Deploy.ProjectName, args.Env)),
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

		functions, err := lambda.NewFunction(ctx, "functions", &lambda.FunctionArgs{
			Code:       args.FunctionsLambda,
			Runtime:    lambda.RuntimeNodeJS20dX,
			MemorySize: pulumi.IntPtr(2048),
			Role:       functionsRole.Arn,
			Handler:    pulumi.String("main.handler"),
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: pulumi.StringMap{
					"KEEL_PROJECT_NAME":      pulumi.String(projectName),
					"KEEL_ENV":               pulumi.String(args.Env),
					"KEEL_FILES_BUCKET_NAME": filesBucket.Bucket,
					// The actual connection string is passed from the runtime to functions
					// via a secret
					"KEEL_DB_CONN_TYPE": pulumi.String("pg"),
					"KEEL_DB_CERT":      pulumi.String("/var/task/rds.pem"),
					"NODE_OPTIONS":      pulumi.String("--enable-source-maps"),
				},
			},
			LoggingConfig: lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("JSON"),
			},
			Tags: baseTags,
		})
		if err != nil {
			return fmt.Errorf("error creating runtime lambda: %v", err)
		}

		runtimeRole, err := createLambdaRole(ctx, "runtime", iam.GetPolicyDocumentStatementArray{
			iam.GetPolicyDocumentStatementInput(iam.GetPolicyDocumentStatementArgs{
				Actions: pulumi.ToStringArray([]string{
					"secretsmanager:GetSecretValue",
				}),
				Resources: pulumi.ToStringArrayOutput([]pulumi.StringOutput{
					dbSecretArn.Elem(),
				}),
			}),
			iam.GetPolicyDocumentStatementInput(iam.GetPolicyDocumentStatementArgs{
				Actions: pulumi.ToStringArray([]string{
					"ssm:GetParameter",
					"ssm:GetParameters",
				}),
				Resources: pulumi.ToStringArray([]string{
					fmt.Sprintf("arn:aws:ssm:%s:%s:parameter%s",
						region,
						args.AwsAccountID,
						SsmParameterName(projectName, args.Env, "*")),
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
		}, baseTags)
		if err != nil {
			return err
		}

		baseRuntimeEnvVars := pulumi.StringMap{
			"KEEL_PROJECT_NAME":        pulumi.String(projectName),
			"KEEL_ENV":                 pulumi.String(args.Env),
			"KEEL_DATABASE_ENDPOINT":   dbInstance.Endpoint,
			"KEEL_DATABASE_DB_NAME":    dbInstance.DbName,
			"KEEL_DATABASE_SECRET_ARN": dbSecretArn.Elem(),
			"KEEL_SECRETS": pulumi.String(strings.Join([]string{
				"KEEL_PRIVATE_KEY",
			}, ":")),
			"KEEL_FILES_BUCKET_NAME": filesBucket.Bucket,
			"KEEL_FUNCTIONS_ARN":     functions.Arn,
			"KEEL_QUEUE_URL":         queue.Url,
		}

		api, err := lambda.NewFunction(ctx, "api", &lambda.FunctionArgs{
			Code:       args.RuntimeLambda,
			Runtime:    lambda.RuntimeCustomAL2023,
			MemorySize: pulumi.IntPtr(2048),
			Role:       runtimeRole.Arn,
			Handler:    pulumi.String("main"),
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: extendStringMap(baseRuntimeEnvVars, pulumi.StringMap{
					"KEEL_RUNTIME_MODE": pulumi.String(RuntimeModeApi),
				}),
			},
			LoggingConfig: lambda.FunctionLoggingConfigArgs{
				LogFormat: pulumi.String("JSON"),
			},
			Tags: baseTags,
		})
		if err != nil {
			return fmt.Errorf("error creating api lambda: %v", err)
		}

		subscriber, err := lambda.NewFunction(ctx, "subscriber", &lambda.FunctionArgs{
			Code:       args.RuntimeLambda,
			Runtime:    lambda.RuntimeCustomAL2023,
			MemorySize: pulumi.IntPtr(2048),
			Role:       runtimeRole.Arn,
			Handler:    pulumi.String("main"),
			Environment: lambda.FunctionEnvironmentArgs{
				Variables: extendStringMap(baseRuntimeEnvVars, pulumi.StringMap{
					"KEEL_RUNTIME_MODE": pulumi.String(RuntimeModeSubscriber),
				}),
			},
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

func SsmParameterName(projectName string, env string, paramName string) string {
	return fmt.Sprintf("/keel/%s/%s/%s", projectName, env, paramName)
}

func createLambdaRole(ctx *pulumi.Context, prefix string, statements iam.GetPolicyDocumentStatementArray, tags pulumi.StringMap) (*iam.Role, error) {
	role, err := iam.NewRole(ctx, fmt.Sprintf("%s-role", prefix), &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				},
				"Effect": "Allow",
				"Sid": ""
			}]
		}`),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating IAM role: %v", err)
	}

	lambdaPolicy, err := iam.NewPolicy(ctx, fmt.Sprintf("%s-policy", prefix), &iam.PolicyArgs{
		Policy: iam.GetPolicyDocumentOutput(ctx, iam.GetPolicyDocumentOutputArgs{
			Statements: statements,
		}).Json(),
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
