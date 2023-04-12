package runtimectx

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamkeel/keel/mail"
)

type mailContextKey string

var mailKey mailContextKey = "mail"

func GetMailClient(ctx context.Context) (*mail.Client, error) {
	v := ctx.Value(mailKey)
	if v == nil {
		return nil, fmt.Errorf("context does not have a :%s key", dbKey)
	}

	client, ok := v.(*mail.Client)

	if !ok {
		return nil, errors.New("mail client in the context has wrong value type")
	}
	return client, nil
}

func WithMailClient(ctx context.Context, client *mail.Client) context.Context {
	return context.WithValue(ctx, mailKey, client)
}
