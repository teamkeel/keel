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

	result, err := database.ExecuteQuery(ctx, input.GetQuery())
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
		if input.GetAfter() != nil && v.StartTime.Before(input.GetAfter().AsTime()) {
			continue
		}

		if input.GetBefore() != nil && v.StartTime.After(input.GetBefore().AsTime()) {
			continue
		}

		if input.Filters != nil {
			filteredOut := false
			for _, f := range input.GetFilters() {
				switch f.GetField() {
				case "error":
					if f.GetValue() == "true" && !v.HasError {
						filteredOut = true
					} else if f.GetValue() == "false" && v.HasError {
						filteredOut = true
					}
				}
				if filteredOut {
					break
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
			// filter out flows API requests
			if strings.Contains(v.RootName, "flows/json") {
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
		return list[i].GetStartTime().AsTime().After(list[j].GetStartTime().AsTime())
	})

	if int(input.GetOffset()) > len(list) {
		list = []*rpc.TraceItem{}
	} else {
		list = list[input.GetOffset():]
	}

	if input.GetLimit() > int32(len(list)) {
		input.Limit = int32(len(list))
	}

	list = list[:input.GetLimit()]

	return &rpc.ListTracesResponse{
		Traces: list,
	}, nil
}

func (s *Server) GetTrace(ctx context.Context, input *rpc.GetTraceRequest) (*rpc.GetTraceResponse, error) {
	trace := localTraceExporter.GetTrace(input.GetTraceId())

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
		Tools:       tools.ActionConfigs(),
		ToolConfigs: tools.GetConfigs(),
	}, nil
}

func (s *Server) ConfigureTool(ctx context.Context, req *rpc.ConfigureToolRequest) (*rpc.ConfigureToolResponse, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	// if we're using the new style of tool configs
	if req.GetToolConfig() != nil {
		updated, err := toolsSvc.ConfigureTool(ctx, req.GetToolConfig())
		if err != nil {
			return nil, twirp.NewError(twirp.Internal, err.Error())
		}
		return &rpc.ConfigureToolResponse{
			ToolConfig: updated,
		}, nil
	}

	// default to old way of configuring tools via ActionConfigs
	updated, err := toolsSvc.ConfigureTool(ctx, req.GetConfiguredTool().ToTool())
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	return &rpc.ConfigureToolResponse{
		ToolConfig: updated,
	}, nil
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
		Tools:       tools.ActionConfigs(),
		ToolConfigs: tools.GetConfigs(),
	}, nil
}

func (s *Server) DuplicateTool(ctx context.Context, req *rpc.DuplicateToolRequest) (*rpc.DuplicateToolResponse, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	newTool, err := toolsSvc.DuplicateTool(ctx, req.GetToolId())
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	return &rpc.DuplicateToolResponse{
		ToolConfig: newTool,
	}, nil
}

// ListFields will list all model & enum fields with their formatting configuration.
func (s *Server) ListFields(ctx context.Context, req *rpc.ListFieldsRequest) (*rpc.ListFieldsResponse, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	fields, err := toolsSvc.GetFields(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}
	if fields == nil {
		return nil, nil
	}

	return &rpc.ListFieldsResponse{
		Fields: fields,
	}, nil
}

// ConfigureFields will configure the formatting of all model & enum fields.
func (s *Server) ConfigureFields(ctx context.Context, req *rpc.ConfigureFieldsRequest) (*rpc.ConfigureFieldsResponse, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	updated, err := toolsSvc.ConfigureFields(ctx, req.GetFields())
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	return &rpc.ConfigureFieldsResponse{
		Fields: updated,
	}, nil
}

// makeToolsService will create a tools service for this server's request (taking in the schema and config from context).
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

func (s *Server) ListPrinters(ctx context.Context, req *rpc.ListPrintersRequest) (*rpc.ListPrintersResponse, error) {
	config, err := GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	resp := rpc.ListPrintersResponse{}
	if config.Hardware != nil {
		for _, p := range config.Hardware.Printers {
			resp.Printers = append(resp.Printers, &rpc.Printer{
				Name: p.Name,
			})
		}
	}

	return &resp, nil
}
