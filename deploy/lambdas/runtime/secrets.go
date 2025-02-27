package runtime

import (
	"context"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

func initSecrets(ctx context.Context, names []string, projectName, env, awsEndpoint string) (map[string]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	opts := []func(*ssm.Options){}
	if awsEndpoint != "" {
		opts = append(opts, func(o *ssm.Options) {
			o.BaseEndpoint = &awsEndpoint
		})
	}

	secretNames := lo.Map(names, func(s string, _ int) string {
		return SsmParameterName(projectName, env, s)
	})

	g, ctx := errgroup.WithContext(ctx)
	params := []types.Parameter{}
	var mutex sync.Mutex

	// Docs for GetParameters state "Maximum number of 10 items"
	// https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_GetParameters.html#API_GetParameters_RequestSyntax
	for _, chunk := range lo.Chunk(secretNames, 10) {
		g.Go(func() error {
			res, err := ssm.NewFromConfig(cfg, opts...).GetParameters(ctx, &ssm.GetParametersInput{
				Names:          chunk,
				WithDecryption: aws.Bool(true),
			})
			if err != nil {
				return err
			}

			mutex.Lock()
			defer mutex.Unlock()
			params = append(params, res.Parameters...)
			return nil
		})
	}

	err = g.Wait()
	if err != nil {
		return nil, err
	}

	secrets := map[string]string{}
	for _, p := range params {
		parts := strings.Split(*p.Name, "/")
		name := parts[len(parts)-1]
		secrets[name] = *p.Value
	}

	return secrets, nil
}
