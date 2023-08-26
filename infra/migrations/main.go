package migrations

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/teamkeel/keel/infra"
	"github.com/teamkeel/keel/migrations"
)

func Start() {
	lambda.Start(handler)
}

func handler(ctx context.Context) error {
	s, err := infra.GetSchema(ctx)
	if err != nil {
		return err
	}

	db, err := infra.GetDatabase(ctx)
	if err != nil {
		return err
	}

	m, err := migrations.New(ctx, s, db)
	if err != nil {
		return err
	}

	return m.Apply(ctx, false)
}
