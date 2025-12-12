package tools

import (
	"context"
	"errors"
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

// ToolsDir is the name of the folder in which tools user config is stored.
const ToolsDir = "tools"

// FieldsFile is the name of the file that holds the user fields configuration for model based config.
const FieldsFile = "_fields.json"

// SpacesFile is the name of the file that holds the spaces configurations.
const SpacesFile = "_spaces.json"

const (
	// Alphabet for unique nanoids to be used for tool ids.
	nanoidABC = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Size of unique nanoids to be used for tool ids.
	nanoidSize = 5
)

type Service struct {
	Schema              *proto.Schema
	Config              *config.ProjectConfig
	ProjectDir          *string
	ToolsConfigStorage  map[string][]byte
	FieldsConfigStorage []byte
	SpacesConfigStorage []byte
}

type ServiceOpt func(s *Service)

func NewService(opts ...ServiceOpt) *Service {
	svc := &Service{}

	for _, o := range opts {
		o(svc)
	}

	return svc
}

func WithSchema(schema *proto.Schema) ServiceOpt {
	return func(s *Service) {
		s.Schema = schema
	}
}

func WithConfig(cfg *config.ProjectConfig) ServiceOpt {
	return func(s *Service) {
		s.Config = cfg
	}
}

// WithFileStorage initialises the tools service with file-baased storage enabled in the given project folder.
func WithFileStorage(projectDir string) ServiceOpt {
	return func(s *Service) {
		s.ProjectDir = &projectDir
	}
}

// WithToolsConfig initialises the tools service with in-memory tools storage and with the given user configuration. This
// option will invalidate any filebased storage that may have been set with `WithFileStorage`.
func WithToolsConfig(store map[string][]byte) ServiceOpt {
	return func(s *Service) {
		s.ToolsConfigStorage = store
		s.ProjectDir = nil
	}
}

// WithFieldsConfig initialises the service with in-memory fields storage and with the given user configuration. This
// option will invalidate any filebased storage that may have been set with `WithFileStorage`.
func WithFieldsConfig(store []byte) ServiceOpt {
	return func(s *Service) {
		s.FieldsConfigStorage = store
		s.ProjectDir = nil
	}
}

// WithSpacesConfig initialises the service with in-memory spaces storage and with the given user configuration. This
// option will invalidate any filebased storage that may have been set with `WithFileStorage`.
func WithSpacesConfig(store []byte) ServiceOpt {
	return func(s *Service) {
		s.SpacesConfigStorage = store
		s.ProjectDir = nil
	}
}

// generateTools will return a map of tool configurations generated for the given schema.
func (s *Service) generateTools(ctx context.Context) ([]*toolsproto.Tool, error) {
	if s.Schema == nil {
		return nil, nil
	}

	gen, err := NewGenerator(s.Schema, s.Config)
	if err != nil {
		return nil, fmt.Errorf("creating tool generator: %w", err)
	}

	if err := gen.Generate(ctx); err != nil {
		return nil, fmt.Errorf("generating tools: %w", err)
	}

	return gen.GetTools(), nil
}

// getGeneratedTool will return the generated tool for the given action/flow name.
func (s *Service) getGeneratedTool(ctx context.Context, name string) (*toolsproto.Tool, error) {
	// generate tools
	genTools, err := s.generateTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating tools from schema: %w", err)
	}

	for _, t := range genTools {
		if t.GetOperationName() == name {
			return t, nil
		}
	}

	return nil, fmt.Errorf("tool not found")
}

// DuplicateTool will take the given tool and duplicate it with a new ID, and then store the changes to files.
func (s *Service) DuplicateTool(ctx context.Context, toolID string) (*toolsproto.Tool, error) {
	tools, err := s.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving tools: %w", err)
	}

	duplicate := tools.FindByID(toolID)
	if duplicate == nil {
		return nil, errors.New("tool not found")
	}

	// generate a unique id suffix
	uid, err := gonanoid.Generate(nanoidABC, nanoidSize)
	if err != nil {
		return nil, fmt.Errorf("generating unique id: %w", err)
	}

	duplicate.Id = casing.ToKebab(duplicate.GetOperationName()) + "-" + uid

	if duplicate.IsActionBased() {
		duplicate.ActionConfig.Name += " (copy)"
		duplicate.ActionConfig.Id = duplicate.GetId()
	} else {
		duplicate.FlowConfig.Name += " (copy)"
	}

	generated, err := s.getGeneratedTool(ctx, duplicate.GetOperationName())
	if err != nil {
		return nil, fmt.Errorf("generating tool %s: %w", duplicate.GetOperationName(), err)
	}

	cfg := extractConfig(generated, duplicate)

	if err := s.addToProject(cfg); err != nil {
		return nil, fmt.Errorf("duplicating tool: %w", err)
	}

	return duplicate, nil
}

// GetTools generates tools based on the schema, reads the configured tools from the project and applies them to the
// generated ones, returning a complete list of tool configs.
func (s *Service) GetTools(ctx context.Context) (*toolsproto.Tools, error) {
	// if we don't have a schema, return nil
	if s.Schema == nil {
		return nil, nil
	}

	// generate tools
	genTools, err := s.generateTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("generating tools from schema: %w", err)
	}

	tools := &toolsproto.Tools{
		Configs: genTools, // all tools, action based and/or flow based
	}

	// load existing configured tools
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading tool configs from file: %w", err)
	}

	// if we have user added tools ...
	if len(userConfig.Tools) > 0 {
		// let's see get the user added tools
		addedIds := tools.DiffIDs(userConfig.Tools.getIDs())
		// for all the added ones, we generate a new tool and add it to our set
		for _, id := range addedIds {
			cfg := userConfig.Tools.findByID(id)
			if cfg == nil {
				continue
			}

			gen, err := s.getGeneratedTool(ctx, cfg.getOperationName())
			if err != nil {
				// cannot generate tool for config, add a blank tool
				if cfg.Type == ToolTypeAction {
					gen = &toolsproto.Tool{
						Type: toolsproto.Tool_ACTION,
						ActionConfig: &toolsproto.ActionConfig{
							ActionName: cfg.getOperationName(),
						},
					}
				} else {
					gen = &toolsproto.Tool{
						Type: toolsproto.Tool_FLOW,
						FlowConfig: &toolsproto.FlowConfig{
							FlowName: cfg.getOperationName(),
						},
					}
				}
			}
			gen.Id = cfg.ID
			if gen.IsActionBased() {
				gen.ActionConfig.Id = cfg.ID
			}

			tools.Configs = append(tools.Configs, gen)
		}
	}

	// if we have user configured field formatting, let's apply them to all the tools
	if len(userConfig.Fields) > 0 {
		userConfig.Fields.applyOnTools(tools.GetConfigs())
	}

	// finally, we now apply all the user configurations to our tools, overwritting any field configs
	if len(userConfig.Tools) > 0 {
		for _, cfg := range userConfig.Tools {
			tool := tools.FindByID(cfg.ID)
			if tool == nil {
				continue
			}

			// apply config on the given tool
			cfg.applyOn(tool)
		}
	}

	NewValidator(s.Schema, tools).validate()

	return tools, nil
}

// ResetTools will remove all the saved tool configs and return the schema generated tools.
func (s *Service) ResetTools(ctx context.Context) (*toolsproto.Tools, error) {
	if err := s.clearTools(); err != nil {
		return nil, fmt.Errorf("removing saved tools configuration: %w", err)
	}

	return s.GetTools(ctx)
}

// ConfigureTool will take the given updated tool config and update the existing project config with it.
func (s *Service) ConfigureTool(ctx context.Context, updated *toolsproto.Tool) (*toolsproto.Tool, error) {
	// get the generated version for the given updated tool
	gen, err := s.getGeneratedTool(ctx, updated.GetOperationName())
	if err != nil {
		return nil, fmt.Errorf("retrieving generated tool: %w", err)
	}

	// load existing configured tools
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading tool configs from file: %w", err)
	}

	// we now apply existing field configuration
	if len(userConfig.Fields) > 0 {
		userConfig.Fields.applyOnTool(gen)
	}

	// we now extract the config from the given tool
	cfg := extractConfig(gen, updated)

	// update the tool saved in project
	err = s.updateToProject(cfg)
	if err != nil {
		return nil, fmt.Errorf("updating tool %s: %w", updated.GetId(), err)
	}

	// read them and return the fresh copy
	tools, err := s.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving tools: %w", err)
	}

	return tools.FindByID(updated.GetId()), nil
}

// GetFields returns the configured fields for this schema.
func (s *Service) GetFields(ctx context.Context) ([]*toolsproto.Field, error) {
	// generate fields
	gen, err := NewGenerator(s.Schema, s.Config)
	if err != nil {
		return nil, fmt.Errorf("creating tool generator: %w", err)
	}

	if err := gen.GenerateFields(ctx); err != nil {
		return nil, fmt.Errorf("generating fields config: %w", err)
	}

	fields := gen.GetFields()

	// load existing configured tools
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading tool configs from file: %w", err)
	}

	// now we apply user config from file
	userConfig.Fields.applyOn(fields)

	return fields, nil
}

// ConfigureFields will take the given updated fields config and update the existing project config with it.
func (s *Service) ConfigureFields(ctx context.Context, updated []*toolsproto.Field) ([]*toolsproto.Field, error) {
	gen, err := NewGenerator(s.Schema, s.Config)
	if err != nil {
		return nil, fmt.Errorf("creating tool generator: %w", err)
	}

	// first we generate fields
	if err := gen.GenerateFields(ctx); err != nil {
		return nil, fmt.Errorf("generating fields config: %w", err)
	}

	// extract user field config by diffing the updated ones with the generated ones
	cfgs := extractFieldsConfigs(gen.GetFields(), updated)

	if err := s.storeFields(cfgs); err != nil {
		return nil, err
	}

	// retrieve newly configured fields
	return s.GetFields(ctx)
}

// GetSpaces returns the configured tool spaces for this project.
func (s *Service) GetSpaces(ctx context.Context) ([]*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading tool configs from file: %w", err)
	}

	return userConfig.Spaces.toProto(), nil
}

// AddSpace will add the given space config to the existing ones and store it.
func (s *Service) AddSpace(ctx context.Context, space *SpaceConfig) (*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading space configs from file: %w", err)
	}

	// set a unique id
	if err := space.setUniqueID(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("creating a unique space id: %w", err)
	}

	userConfig.Spaces = append(userConfig.Spaces, space)

	if err := s.storeSpaces(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("storing space configs: %w", err)
	}

	return space.toProto(), nil
}

// UpdateSpace will update an existing space with the given data (note that containing space items are not subject to updates).
func (s *Service) UpdateSpace(ctx context.Context, updated *SpaceConfig) (*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading space configs from file: %w", err)
	}

	space := userConfig.Spaces.findByID(updated.ID)
	if space == nil {
		return nil, nil
	}

	space.Icon = updated.Icon
	space.Name = updated.Name
	space.DisplayOrder = updated.DisplayOrder

	if err := s.storeSpaces(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("storing space configs: %w", err)
	}

	return space.toProto(), nil
}

// RemoveSpace will remove the given space config from the storage.
func (s *Service) RemoveSpace(ctx context.Context, spaceID string) error {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return fmt.Errorf("loading space configs from file: %w", err)
	}

	remainingSpaces := SpaceConfigs{}

	for _, sp := range userConfig.Spaces {
		if sp.ID != spaceID {
			remainingSpaces = append(remainingSpaces, sp)
		}
	}

	return s.storeSpaces(remainingSpaces)
}
func (s *Service) RemoveSpaceItem(ctx context.Context, spaceID, itemID string) (*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading space configs from file: %w", err)
	}

	space := userConfig.Spaces.findByID(spaceID)
	if space == nil {
		return nil, fmt.Errorf("space not found")
	}

	space.removeItem(itemID)

	if err := s.storeSpaces(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("storing space configs: %w", err)
	}

	return space.toProto(), nil
}

// AddSpaceAction adds an action to the given space. If Group ID is not empty, the action will be added to the given group.
func (s *Service) AddSpaceAction(ctx context.Context, payload *toolsproto.CreateSpaceActionPayload) (*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading space configs from file: %w", err)
	}

	space := userConfig.Spaces.findByID(payload.GetSpaceId())
	if space == nil {
		return nil, fmt.Errorf("space not found")
	}

	action := SpaceAction{
		Link: extractLinkConfig(nil, payload.GetLink()),
	}

	if action.Link == nil {
		return nil, fmt.Errorf("invalid link")
	}

	// set a unique ID
	if err := action.setUniqueID(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("generating unique id: %w", err)
	}

	if err := space.addAction(&action, payload.GetGroupId()); err != nil {
		return nil, fmt.Errorf("adding an action: %w", err)
	}

	if err := s.storeSpaces(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("storing space configs: %w", err)
	}

	return space.toProto(), nil
}

func (s *Service) AddSpaceGroup(ctx context.Context, payload *toolsproto.CreateSpaceGroupPayload) (*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading space configs from file: %w", err)
	}

	space := userConfig.Spaces.findByID(payload.GetSpaceId())
	if space == nil {
		return nil, fmt.Errorf("space not found")
	}

	group := SpaceGroup{
		Name: func() string {
			if payload.GetName() != nil {
				return payload.GetName().GetTemplate()
			}

			return ""
		}(),
		Description: func() string {
			if payload.GetDescription() != nil {
				return payload.GetDescription().GetTemplate()
			}

			return ""
		}(),
		DisplayOrder: payload.GetDisplayOrder(),
	}

	// set a unique ID
	if err := group.setUniqueID(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("generating unique id: %w", err)
	}

	space.Groups = append(space.Groups, &group)

	if err := s.storeSpaces(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("storing space configs: %w", err)
	}

	return space.toProto(), nil
}

func (s *Service) AddSpaceMetric(ctx context.Context, payload *toolsproto.CreateSpaceMetricPayload) (*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading space configs from file: %w", err)
	}

	space := userConfig.Spaces.findByID(payload.GetSpaceId())
	if space == nil {
		return nil, fmt.Errorf("space not found")
	}

	metric := SpaceMetric{
		Label: func() string {
			if payload.GetLabel() != nil {
				return payload.GetLabel().GetTemplate()
			}

			return ""
		}(),
		ToolID:        payload.GetToolId(),
		DisplayOrder:  payload.GetDisplayOrder(),
		FacetLocation: payload.GetFacetLocation().GetPath(),
	}

	// set a unique ID
	if err := metric.setUniqueID(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("generating unique id: %w", err)
	}

	space.Metrics = append(space.Metrics, &metric)

	if err := s.storeSpaces(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("storing space configs: %w", err)
	}

	return space.toProto(), nil
}

func (s *Service) AddSpaceLink(ctx context.Context, payload *toolsproto.CreateSpaceLinkPayload) (*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading space configs from file: %w", err)
	}

	space := userConfig.Spaces.findByID(payload.GetSpaceId())
	if space == nil {
		return nil, fmt.Errorf("space not found")
	}

	link := SpaceLink{
		Link: extractExternalLink(payload.GetLink()),
	}

	// set a unique ID
	if err := link.setUniqueID(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("generating unique id: %w", err)
	}

	space.Links = append(space.Links, &link)

	if err := s.storeSpaces(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("storing space configs: %w", err)
	}

	return space.toProto(), nil
}

func (s *Service) UpdateSpaceItem(ctx context.Context, protoPayload any) (*toolsproto.Space, error) {
	// load existing user configuration
	userConfig, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading space configs from file: %w", err)
	}

	spaceID := ""
	for _, space := range userConfig.Spaces {
		switch payload := protoPayload.(type) {
		case *toolsproto.UpdateSpaceActionPayload:
			if item := space.allActions().findByID(payload.GetId()); item != nil {
				item.Link = extractLinkConfig(nil, payload.GetLink())
				spaceID = space.ID
			}
		case *toolsproto.UpdateSpaceMetricPayload:
			if item := space.Metrics.findByID(payload.GetId()); item != nil {
				item.Label = func() string {
					if payload.GetLabel() != nil {
						return payload.GetLabel().GetTemplate()
					}

					return ""
				}()
				item.ToolID = payload.GetToolId()
				item.DisplayOrder = payload.GetDisplayOrder()
				item.FacetLocation = payload.GetFacetLocation().GetPath()
				spaceID = space.ID
			}
		case *toolsproto.UpdateSpaceGroupPayload:
			if item := space.Groups.findByID(payload.GetId()); item != nil {
				item.Name = func() string {
					if payload.GetName() != nil {
						return payload.GetName().GetTemplate()
					}

					return ""
				}()
				item.Description = func() string {
					if payload.GetDescription() != nil {
						return payload.GetDescription().GetTemplate()
					}

					return ""
				}()
				item.DisplayOrder = payload.GetDisplayOrder()
				spaceID = space.ID
			}
		case *toolsproto.UpdateSpaceLinkPayload:
			if item := space.Links.findByID(payload.GetId()); item != nil {
				item.Link = extractExternalLink(payload.GetLink())
				spaceID = space.ID
			}
		}
	}

	space := userConfig.Spaces.findByID(spaceID)

	if err := s.storeSpaces(userConfig.Spaces); err != nil {
		return nil, fmt.Errorf("storing space configs: %w", err)
	}

	return space.toProto(), nil
}
