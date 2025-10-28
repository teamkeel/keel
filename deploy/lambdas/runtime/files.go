package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/vincent-petithory/dataurl"
	"go.opentelemetry.io/otel/trace"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	endpoints "github.com/aws/smithy-go/endpoints"
	"github.com/teamkeel/keel/storage"
)

const (
	FileObjectExpiryDuration = time.Hour

	// Note: it's important this matches the prefix used in the functions-runtime so that
	// files uploaded here can be read by functions and vice-versa.
	FileObjectPrefix = "files/"
)

var _ storage.Storer = &S3BucketStore{}

type S3BucketStore struct {
	// TODO: storing a context like this is an anti-pattern in Go - the methods in storage.Storer should all take context instead
	context context.Context

	client     *s3.Client
	bucketName string
	tracer     trace.Tracer
}

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

func initFiles(ctx context.Context, tracer trace.Tracer, bucketName, awsEndpoint string) (*S3BucketStore, error) {
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

	return &S3BucketStore{
		context:    ctx,
		tracer:     tracer,
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s S3BucketStore) GetFileInfo(key string) (storage.FileInfo, error) {
	if s.bucketName == "" {
		return storage.FileInfo{}, errors.New("S3 bucket name cannot be empty")
	}

	pathedKey := path.Join(FileObjectPrefix, key)

	object, err := s.client.GetObject(s.context, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &pathedKey})
	if err != nil {
		return storage.FileInfo{}, err
	}

	return storage.FileInfo{
		Key:         key,
		Filename:    object.Metadata["filename"],
		ContentType: *object.ContentType,
		Size:        int(*object.ContentLength),
	}, nil
}

func (s S3BucketStore) GetFileData(key string) ([]byte, storage.FileInfo, error) {
	if s.bucketName == "" {
		return nil, storage.FileInfo{}, errors.New("S3 bucket name cannot be empty")
	}

	pathedKey := path.Join(FileObjectPrefix, key)

	object, err := s.client.GetObject(s.context, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &pathedKey})
	if err != nil {
		return nil, storage.FileInfo{}, err
	}

	defer object.Body.Close()

	// Read contents into memory
	data, err := io.ReadAll(object.Body)
	if err != nil {
		return nil, storage.FileInfo{}, err
	}

	return data, storage.FileInfo{
		Key:         key,
		Filename:    object.Metadata["filename"],
		ContentType: *object.ContentType,
		Size:        int(*object.ContentLength),
	}, nil
}

func (s S3BucketStore) Store(dataURL string) (storage.FileInfo, error) {
	var span trace.Span
	s.context, span = s.tracer.Start(s.context, "Store File")
	defer span.End()

	if s.bucketName == "" {
		return storage.FileInfo{}, errors.New("S3 bucket name cannot be empty")
	}

	durl, err := dataurl.DecodeString(dataURL)
	if err != nil {
		return storage.FileInfo{}, fmt.Errorf("decoding dataurl: %w", err)
	}

	key := ksuid.New().String()
	pathedKey := path.Join(FileObjectPrefix, key)
	ct := durl.ContentType()

	_, err = s.client.PutObject(s.context, &s3.PutObjectInput{
		Bucket:      &s.bucketName,
		Key:         &pathedKey,
		Body:        bytes.NewReader(durl.Data),
		ContentType: &ct,
		Metadata:    map[string]string{"filename": durl.Params["name"]}})
	if err != nil {
		return storage.FileInfo{}, fmt.Errorf("storing file: %w", err)
	}

	return s.GetFileInfo(key)
}

func (s S3BucketStore) GenerateFileResponse(fi *storage.FileInfo) (storage.FileResponse, error) {
	var span trace.Span
	s.context, span = s.tracer.Start(s.context, "Hydrate File")
	defer span.End()

	if s.bucketName == "" {
		return storage.FileResponse{}, errors.New("S3 bucket name cannot be empty")
	}

	hydrated, err := s.getPresignedURL(fi)
	if err != nil {
		return storage.FileResponse{}, fmt.Errorf("hydrating file info: %w", err)
	}

	return hydrated, nil
}

func (s S3BucketStore) getPresignedURL(fi *storage.FileInfo) (storage.FileResponse, error) {
	if s.bucketName == "" {
		return storage.FileResponse{}, errors.New("S3 bucket name cannot be empty")
	}

	pathedKey := path.Join(FileObjectPrefix, fi.Key)

	presignClient := s3.NewPresignClient(s.client)
	request, err := presignClient.PresignGetObject(s.context, &s3.GetObjectInput{
		Bucket:                     &s.bucketName,
		Key:                        &pathedKey,
		ResponseContentDisposition: aws.String(fmt.Sprintf(`attachment; filename="%s"`, fi.Filename)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = FileObjectExpiryDuration
	})
	if err != nil {
		return storage.FileResponse{}, fmt.Errorf("couldn't get a presigned url for %s:%s. %w", s.bucketName, pathedKey, err)
	}

	return storage.FileResponse{
		Key:         fi.Key,
		Filename:    fi.Filename,
		ContentType: fi.ContentType,
		Size:        fi.Size,
		URL:         request.URL,
	}, nil
}
