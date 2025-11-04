package runtime

import (
	"context"
	"net/url"

	"go.opentelemetry.io/otel/trace"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	endpoints "github.com/aws/smithy-go/endpoints"
	"github.com/teamkeel/keel/storage"
)

// CustomS3EndpointResolverV2 allows us to use a custom endpoint.
//
// If a custom endpoint is set we need to use a custom resolver. Just settng the base endpoint isn't enough for S3
// as the default resolver uses the bucket name as a sub-domain, which likely won't work with the custom endpoint.
// By implementing a full resolver we can force it to be the endpoint we want.
type CustomS3EndpointResolverV2 struct {
	endpoint string
}

func (e *CustomS3EndpointResolverV2) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (endpoints.Endpoint, error) {
	v, err := url.Parse(e.endpoint)
	if err != nil {
		return endpoints.Endpoint{}, err
	}

	return endpoints.Endpoint{
		URI: *v,
	}, nil
}

func initFiles(ctx context.Context, tracer trace.Tracer, bucketName, awsEndpoint string) (*storage.S3BucketStore, error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, err
	}

	opts := []func(*s3.Options){}
	if awsEndpoint != "" {
		opts = append(opts, s3.WithEndpointResolverV2(&CustomS3EndpointResolverV2{
			endpoint: awsEndpoint,
		}))
	}

	client := s3.NewFromConfig(cfg, opts...)

	return storage.NewS3BucketStore(ctx, bucketName, client, tracer), nil
}
