package rpcApi

import (
	"context"

	"github.com/teamkeel/keel/tools"

	"github.com/teamkeel/keel/rpc/rpc"
	"github.com/twitchtv/twirp"
)

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

// List all the tool spaces defined in this repo.
func (s *Server) ListToolSpaces(ctx context.Context, req *rpc.ListToolSpacesRequest) (*rpc.ListToolSpacesResponse, error) {
	toolsSvc, err := s.makeToolsService(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}

	spaces, err := toolsSvc.GetSpaces(ctx)
	if err != nil {
		return nil, twirp.NewError(twirp.Internal, err.Error())
	}
	if spaces == nil {
		return nil, nil
	}

	return &rpc.ListToolSpacesResponse{
		Spaces: spaces,
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

// Creates a new tool space. Returns the newly added space.
func (s *Server) CreateToolSpace(ctx context.Context, req *rpc.CreateToolSpaceRequest) (*rpc.ToolSpaceResponse, error) {
	//TODO: implement
	return nil, nil
}

// Removes a space from the project.
func (s *Server) RemoveToolSpace(ctx context.Context, req *rpc.RemoveToolSpaceRequest) (*rpc.RemoveToolSpaceResponse, error) {
	//TODO: implement
	return nil, nil
}

// Adds a new space item (group, action, metric) to a given space. Returns the updated space.
func (s *Server) AddToolSpaceItem(ctx context.Context, req *rpc.AddToolSpaceItemRequest) (*rpc.ToolSpaceResponse, error) {
	//TODO: implement
	return nil, nil
}

// Removes a space item from a given space. Returns the updated space.
func (s *Server) RemoveToolSpaceItem(ctx context.Context, req *rpc.RemoveToolSpaceItemRequest) (*rpc.ToolSpaceResponse, error) {
	//TODO: implement
	return nil, nil
}

// Updates a tool space item (group, action, metric). Returns the updated space.
func (s *Server) UpdateToolSpaceItem(ctx context.Context, req *rpc.UpdateToolSpaceItemRequest) (*rpc.ToolSpaceResponse, error) {
	//TODO: implement
	return nil, nil
}
