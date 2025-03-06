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
	"github.com/teamkeel/keel/tools"
	toolsproto "github.com/teamkeel/keel/tools/proto"

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

	b, err := json.Marshal(result)
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

	verbose := GetTraceVerbosity(ctx)

	list := []*rpc.TraceItem{}

	for k, v := range traces {
		if input.After != nil && v.StartTime.Before(input.After.AsTime()) {
			continue
		}

		if input.Before != nil && v.StartTime.After(input.Before.AsTime()) {
			continue
		}

		if input.Filters != nil {
			filteredOut := false
			for _, f := range input.Filters {
				switch f.Field {
				case "error":
					if f.Value == "true" && !v.HasError {
						filteredOut = true
						break
					} else if f.Value == "false" && v.HasError {
						filteredOut = true
						break
					}
				}
			}
			if filteredOut {
				continue
			}
		}

		if !verbose {
			if v.Type != "request" {
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

func (s *Server) ListTools(ctx context.Context, input *rpc.ListToolsRequest) (*rpc.ListToolsResponse, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	tools, err := toolsSvc.GetTools(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}
	if tools == nil {
		return nil, nil
	}
	return &rpc.ListToolsResponse{
		Tools: tools.Tools,
	}, nil
}

func (s *Server) ConfigureTool(ctx context.Context, req *rpc.ConfigureToolRequest) (*toolsproto.ActionConfig, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	updated, err := toolsSvc.ConfigureTool(ctx, req.GetConfiguredTool())
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	return updated, nil
}

func (s *Server) ResetTools(ctx context.Context, req *rpc.ResetToolsRequest) (*rpc.ResetToolsResponse, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	tools, err := toolsSvc.ResetTools(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	return &rpc.ResetToolsResponse{
		Tools: tools.Tools,
	}, nil
}

func (s *Server) DuplicateTool(ctx context.Context, req *rpc.DuplicateToolRequest) (*toolsproto.ActionConfig, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	newTool, err := toolsSvc.DuplicateTool(ctx, req.GetToolId())
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	return newTool, nil
}

// makeToolsService will create a tools service for this server's request (taking in the schema and config from context)
func (s *Server) makeToolsService(ctx context.Context) (*tools.Service, error) {
	schema, err := GetSchema(ctx)
	if err != nil {
		return nil, err
	}

	config, err := GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	projectDir, err := GetProjectDir(ctx)
	if err != nil {
		return nil, err
	}

	return tools.NewService(tools.WithFileStorage(projectDir), tools.WithConfig(config), tools.WithSchema(schema)), nil
}
