package rpcApi

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/teamkeel/keel/cmd/localTraceExporter"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/rpc/rpc"
	"github.com/twitchtv/twirp"
	v1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct{}

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
	traces := localTraceExporter.Summary()

	list := []*rpc.TraceItem{}

	for k, v := range traces {

		if input.After != nil && v.StartTime.Before(input.After.AsTime()) {
			continue
		}

		if input.Before != nil && v.StartTime.After(input.Before.AsTime()) {
			continue
		}

		if v.RootName == "GET /_health" {
			continue
		}
		if strings.HasSuffix(v.RootName, "openapi.json") {
			continue
		}
		if strings.HasPrefix(v.RootName, "OPTIONS") {
			continue
		}

		list = append(list, &rpc.TraceItem{
			TraceId:    k,
			RootName:   v.RootName,
			StartTime:  timestamppb.New(v.StartTime),
			EndTime:    timestamppb.New(v.EndTime),
			DurationMs: float32(v.Duration.Milliseconds()),
			Error:      v.HasError,
		})
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].StartTime.AsTime().After(list[j].StartTime.AsTime())
	})

	if int(input.Offset) > len(list) {
		list = []*rpc.TraceItem{}
	} else {
		list = list[input.Offset:]
	}

	if input.Limit > int32(len(list)) {
		input.Limit = int32(len(list))
	}

	list = list[:input.Limit]

	return &rpc.ListTracesResponse{
		Traces: list,
	}, nil
}

func (s *Server) GetTrace(ctx context.Context, input *rpc.GetTraceRequest) (*rpc.GetTraceResponse, error) {
	trace := localTraceExporter.GetTrace(input.TraceId)

	if trace == nil {
		return nil, twirp.NewError(twirp.NotFound, "trace not found")
	}

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
