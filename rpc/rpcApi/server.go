package rpcApi

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/teamkeel/keel/cmd/localTraceExporter"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/rpc/rpc"
	"github.com/twitchtv/twirp"
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
)

type Server struct{}

type schemaContextKey string

var schemaKey schemaContextKey = "schema"

func GetSchema(ctx context.Context) (*proto.Schema, error) {
	v := ctx.Value(schemaKey)
	schema, ok := v.(*proto.Schema)

	if !ok {
		return nil, errors.New("database in the context has wrong value type")
	}
	return schema, nil
}

func WithSchema(ctx context.Context, schema *proto.Schema) context.Context {
	return context.WithValue(ctx, schemaKey, schema)
}

func (s *Server) GetActiveSchema(ctx context.Context, req *rpc.GetSchemaRequest) (*rpc.GetSchemaResponse, error) {
	schema, err := GetSchema(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	if schema == nil {
		schema = &proto.Schema{}
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
		}, nil
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

func (s *Server) ListTraces(ctx context.Context, input *rpc.ListTracesRequest) (*rpc.ListTracesResponse, error) {
	traces := localTraceExporter.AllTraces()

	tracesSlice := []*rpc.TraceItem{}

	for k, _ := range traces {
		tracesSlice = append(tracesSlice, &rpc.TraceItem{
			TraceId: k,
		})
	}

	return &rpc.ListTracesResponse{
		Traces: tracesSlice,
	}, nil
}

func (s *Server) GetTrace(ctx context.Context, input *rpc.GetTraceRequest) (*rpc.GetTraceResponse, error) {
	trace := localTraceExporter.GetTrace(input.TraceId)

	traceData := v1.TracesData{
		ResourceSpans: []*v1.ResourceSpans{
			{
				ScopeSpans: []*v1.ScopeSpans{
					{Spans: trace},
				},
			},
		},
	}

	return &rpc.GetTraceResponse{
		Trace: &traceData,
	}, nil
}
