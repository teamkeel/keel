package rpcApiServer

import (
	"context"
	"encoding/json"

	"github.com/teamkeel/keel/db"
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

func (s *Server) RunSQLQuery(ctx context.Context, input *rpc.SQLQueryInput) (*rpc.SQLQueryResponse, error) {
	database, err := db.GetDatabase(ctx)
	if err != nil {
		return &rpc.SQLQueryResponse{
			Status: rpc.SQLQueryStatus_failed,
			Error:  err.Error(),
		}, err
	}

	result, err := database.ExecuteQuery(ctx, input.Query)
	if err != nil {
		return &rpc.SQLQueryResponse{
			Status: rpc.SQLQueryStatus_failed,
			Error:  err.Error(),
		}, err
	}

	b, err := json.Marshal(result.Rows)
	if err != nil {
		return &rpc.SQLQueryResponse{
			Status: rpc.SQLQueryStatus_failed,
			Error:  err.Error(),
		}, err
	}

	return &rpc.SQLQueryResponse{
		Status:      rpc.SQLQueryStatus_success,
		ResultsJSON: string(b),
		TotalRows:   int32(len(result.Rows)),
	}, nil
}
