package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/vincent-petithory/dataurl"
	"go.opentelemetry.io/otel/trace"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/teamkeel/keel/storage"
)

const (
	FileObjectExpiryDuration = time.Hour
)

var _ storage.Storer = &S3BucketStore{}

type S3BucketStore struct {
	context    context.Context
	client     *s3.Client
	bucketName string
}

func NewS3BucketStore(ctx context.Context) *S3BucketStore {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3BucketStore{
		context:    ctx,
		client:     client,
		bucketName: os.Getenv("KEEL_FILES_BUCKET_NAME"),
	}
}

func (s S3BucketStore) GetFileInfo(key string) (storage.FileInfo, error) {
	if s.bucketName == "" {
		return storage.FileInfo{}, errors.New("S3 bucket name cannot be empty")
	}

	object, err := s.client.GetObject(s.context, &s3.GetObjectInput{
		Bucket: &s.bucketName,
		Key:    &key})
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

func (s S3BucketStore) Store(dataURL string) (storage.FileInfo, error) {
	var span trace.Span
	s.context, span = tracer.Start(s.context, "Store File")
	defer span.End()

	if s.bucketName == "" {
		return storage.FileInfo{}, errors.New("S3 bucket name cannot be empty")
	}

	durl, err := dataurl.DecodeString(dataURL)
	if err != nil {
		return storage.FileInfo{}, fmt.Errorf("decoding dataurl: %w", err)
	}

	key := ksuid.New().String()
	ct := durl.ContentType()

	_, err = s.client.PutObject(s.context, &s3.PutObjectInput{
		Bucket:      &s.bucketName,
		Key:         &key,
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
	s.context, span = tracer.Start(s.context, "Hydrate File")
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

	presignClient := s3.NewPresignClient(s.client)
	request, err := presignClient.PresignGetObject(s.context, &s3.GetObjectInput{
		Bucket:                     &s.bucketName,
		Key:                        &fi.Key,
		ResponseContentDisposition: aws.String(fmt.Sprintf(`attachment; filename="%s"`, fi.Filename)),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = FileObjectExpiryDuration
	})
	if err != nil {
		return storage.FileResponse{}, fmt.Errorf("couldn't get a presigned url for %s:%s. %w", s.bucketName, fi.Key, err)
	}

	return storage.FileResponse{
		Key:         fi.Key,
		Filename:    fi.Filename,
		ContentType: fi.ContentType,
		Size:        fi.Size,
		URL:         request.URL,
	}, nil
}
