package rpcApiServer

import (
	"context"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/rpc/rpc"
	"github.com/twitchtv/twirp"
)

type Server struct{}

func (s *Server) GetActiveSchema(ctx context.Context, req *rpc.GetSchemaRequest) (*rpc.GetSchemaResponse, error) {

	schema, ok := ctx.Value("schema").(*proto.Schema)
	if !ok {
		return nil, twirp.NewError(twirp.NotFound, "schema not found")
	}

	return &rpc.GetSchemaResponse{
		Schema: schema,
	}, nil
}

func (s *Server) Ping(context.Context, *rpc.PingRequest) (*rpc.PingResponse, error) {
	return &rpc.PingResponse{
		Message: "pong",
	}, nil
}
