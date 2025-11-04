package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/segmentio/ksuid"
	"github.com/vincent-petithory/dataurl"
	"go.opentelemetry.io/otel/trace"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	FileObjectExpiryDuration = time.Duration(60 * time.Minute)
	FileObjectPrefix         = "files/"
)

var _ Storer = &S3BucketStore{}

type S3BucketStore struct {
	context    context.Context
	tracer     trace.Tracer
	Client     *s3.Client
	BucketName string
}

func NewS3BucketStore(ctx context.Context, bucketName string, client *s3.Client, tracer trace.Tracer) *S3BucketStore {
	return &S3BucketStore{
		context:    ctx,
		Client:     client,
		BucketName: bucketName,
		tracer:     tracer,
	}
}

func (s S3BucketStore) GetFileInfo(key string) (FileInfo, error) {
	if s.BucketName == "" {
		return FileInfo{}, errors.New("S3 bucket name cannot be empty")
	}

	pathedKey := path.Join(FileObjectPrefix, key)

	object, err := s.Client.GetObject(s.context, &s3.GetObjectInput{
		Bucket: &s.BucketName,
		Key:    &pathedKey})
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Key:         key,
		Filename:    object.Metadata["filename"],
		ContentType: *object.ContentType,
		Size:        int(*object.ContentLength),
	}, nil
}

func (s S3BucketStore) Store(dataURL string) (FileInfo, error) {
	var span trace.Span
	s.context, span = s.tracer.Start(s.context, "Store File")
	defer span.End()

	if s.BucketName == "" {
		return FileInfo{}, errors.New("S3 bucket name cannot be empty")
	}

	durl, err := dataurl.DecodeString(dataURL)
	if err != nil {
		return FileInfo{}, fmt.Errorf("decoding dataurl: %w", err)
	}

	key := ksuid.New().String()
	pathedKey := path.Join(FileObjectPrefix, key)
	ct := durl.ContentType()

	_, err = s.Client.PutObject(s.context, &s3.PutObjectInput{
		Bucket:      &s.BucketName,
		Key:         &pathedKey,
		Body:        bytes.NewReader(durl.Data),
		ContentType: &ct,
		Metadata:    map[string]string{"filename": durl.Params["name"]}})
	if err != nil {
		return FileInfo{}, fmt.Errorf("storing file: %w", err)
	}

	return s.GetFileInfo(key)
}

func (s S3BucketStore) GenerateFileResponse(fi *FileInfo) (FileResponse, error) {
	var span trace.Span
	s.context, span = s.tracer.Start(s.context, "Hydrate File")
	defer span.End()

	if s.BucketName == "" {
		return FileResponse{}, errors.New("S3 bucket name cannot be empty")
	}

	hydrated, err := s.getPresignedURL(fi)
	if err != nil {
		return FileResponse{}, fmt.Errorf("hydrating file info: %w", err)
	}

	return hydrated, nil
}

func (s S3BucketStore) GetFileData(key string) ([]byte, FileInfo, error) {
	if s.BucketName == "" {
		return nil, FileInfo{}, errors.New("S3 bucket name cannot be empty")
	}

	pathedKey := path.Join(FileObjectPrefix, key)

	object, err := s.Client.GetObject(s.context, &s3.GetObjectInput{
		Bucket: &s.BucketName,
		Key:    &pathedKey})
	if err != nil {
		return nil, FileInfo{}, err
	}

	defer object.Body.Close()

	// Read contents into memory
	data, err := io.ReadAll(object.Body)
	if err != nil {
		return nil, FileInfo{}, err
	}

	return data, FileInfo{
		Key:         key,
		Filename:    object.Metadata["filename"],
		ContentType: *object.ContentType,
		Size:        int(*object.ContentLength),
	}, nil
}

func (s S3BucketStore) getPresignedURL(fi *FileInfo) (FileResponse, error) {
	if s.BucketName == "" {
		return FileResponse{}, errors.New("S3 bucket name cannot be empty")
	}

	pathedKey := path.Join(FileObjectPrefix, fi.Key)

	presignClient := s3.NewPresignClient(s.Client)

	request, err := presignClient.PresignGetObject(s.context, &s3.GetObjectInput{
		Bucket:                     &s.BucketName,
		Key:                        &pathedKey,
		ResponseContentDisposition: aws.String("inline"),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = FileObjectExpiryDuration
	})
	if err != nil {
		return FileResponse{}, fmt.Errorf("couldn't get a presigned url for %s:%s. %w", s.BucketName, pathedKey, err)
	}

	return FileResponse{
		Key:         fi.Key,
		Filename:    fi.Filename,
		ContentType: fi.ContentType,
		Size:        fi.Size,
		URL:         request.URL,
	}, nil
}
