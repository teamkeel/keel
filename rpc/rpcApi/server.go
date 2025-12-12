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

func (s *Server) ListPrinters(ctx context.Context, req *rpc.ListPrintersRequest) (*rpc.ListPrintersResponse, error) {
	config, err := GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	resp := rpc.ListPrintersResponse{}
	if config != nil && config.Hardware != nil {
		for _, p := range config.Hardware.Printers {
			resp.Printers = append(resp.Printers, &rpc.Printer{
				Name: p.Name,
			})
		}
	}

	return &resp, nil
}
