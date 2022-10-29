package rpc

import (
	"github.com/teamkeel/keel/runtime/actions"
)

type RpcArgParser struct {
}

// TODO: In here we must inspect the data structures and parse arguments accordingly.  E.g. for date and time to time.Time

func (parser *RpcArgParser) ParseGet(requestInput map[string]any) (*actions.Args, error) {
	return actions.NewArgs(map[string]any{}, map[string]any{}), nil
}

func (parser *RpcArgParser) ParseCreate(requestInput map[string]any) (*actions.Args, error) {
	return actions.NewArgs(map[string]any{}, map[string]any{}), nil
}

func (parser *RpcArgParser) ParseUpdate(requestInput map[string]any) (*actions.Args, error) {
	return actions.NewArgs(map[string]any{}, map[string]any{}), nil
}

func (parser *RpcArgParser) ParseList(requestInput map[string]any) (*actions.Args, error) {
	return actions.NewArgs(map[string]any{}, map[string]any{}), nil
}

func (parser *RpcArgParser) ParseDelete(requestInput map[string]any) (*actions.Args, error) {
	return actions.NewArgs(map[string]any{}, map[string]any{}), nil
}
