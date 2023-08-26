package infra

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/sst-go"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	schemaMutex sync.Mutex
	schema      *proto.Schema
)

func GetSchema(ctx context.Context) (*proto.Schema, error) {
	if schema != nil {
		return schema, nil
	}

	schemaMutex.Lock()
	defer schemaMutex.Unlock()

	bucket, err := sst.Bucket(ctx, "RuntimeAssets")
	if err != nil {
		return nil, err
	}

	key := os.Getenv("KEEL_SCHEMA_FILE_KEY")
	if key == "" {
		return nil, errors.New("missing env var KEEL_SCHEMA_FILE_KEY")
	}

	b, err := GetS3File(ctx, bucket.BucketName, key)
	if err != nil {
		return nil, err
	}

	var s proto.Schema
	err = protojson.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}

	schema = &s
	return schema, nil
}
