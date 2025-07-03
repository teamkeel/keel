package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/teamkeel/keel/casing"
	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

const toolsDir = "tools"

const (
	// Alphabet for unique nanoids to be used for tool ids.
	nanoidABC = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Size of unique nanoids to be used for tool ids.
	nanoidSize = 5
)

type Service struct {
	Schema             *proto.Schema
	Config             *config.ProjectConfig
	ProjectDir         *string
	ToolsConfigStorage map[string][]byte
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

// storageFolder return the path to the tools config storage folder.
func (s *Service) storageFolder() string {
	if s.ProjectDir == nil {
		return ""
	}

	return filepath.Join(*s.ProjectDir, toolsDir)
}

// hasFileStorage tells us if this service has been initialised with file storage.
func (s *Service) hasFileStorage() bool {
	return s.ProjectDir != nil
}

func (s *Service) initStorageFolder() error {
	if !s.hasFileStorage() {
		return nil
	}

	if _, err := os.Stat(s.storageFolder()); err != nil {
		err := os.Mkdir(s.storageFolder(), os.ModePerm)
		if err != nil {
			return fmt.Errorf("initialising tools dir: %w", err)
		}
	}
	return nil
}

// loadFromFileStorage will load configs from file storage.
func (s *Service) loadFromFileStorage() (ToolConfigs, error) {
	if !s.hasFileStorage() {
		return nil, fmt.Errorf("service does not have file storage enabled")
	}

	if err := s.initStorageFolder(); err != nil {
		return nil, fmt.Errorf("initialising tools folder: %w", err)
	}

	configFiles, err := filepath.Glob(filepath.Join(s.storageFolder(), "*.json"))
	if err != nil {
		return nil, err
	}

	cfgs := ToolConfigs{}

	for _, fName := range configFiles {
		fileBytes, err := os.ReadFile(fName)
		if err != nil {
			return nil, err
		}
		var cfg ToolConfig
		if err := json.Unmarshal(fileBytes, &cfg); err != nil {
			return nil, err
		}
		cfgs = append(cfgs, &cfg)
	}

	return cfgs, nil
}

// load will read the tools configurations from storage.
func (s *Service) load() (ToolConfigs, error) {
	// if we have file storage enabled, load from file
	if s.hasFileStorage() {
		return s.loadFromFileStorage()
	}

	// otherwise load from internal cache
	cfgs := ToolConfigs{}
	for _, fileBytes := range s.ToolsConfigStorage {
		var cfg ToolConfig
		if err := json.Unmarshal(fileBytes, &cfg); err != nil {
			return nil, err
		}
		cfgs = append(cfgs, &cfg)
	}

	return cfgs, nil
}

// clear will remove all the saved tool configs from the project.
func (s *Service) clear() error {
	if s.hasFileStorage() {
		err := os.RemoveAll(s.storageFolder())
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil
			}
			return err
		}
		return nil
	}

	s.ToolsConfigStorage = map[string][]byte{}
	return nil
}

// store will save the given tools configuration to the tools.json file in the project.
func (s *Service) store(cfgs ToolConfigs) error {
	if s.hasFileStorage() {
		if err := s.initStorageFolder(); err != nil {
			return fmt.Errorf("initialising tools folder: %w", err)
		}

		for _, cfg := range cfgs {
			if !cfg.hasChanges() {
				// no changes to this tool, so remove any existing config for this tool
				if err := os.Remove(filepath.Join(s.storageFolder(), cfg.ID+".json")); err != nil {
					if !errors.Is(err, os.ErrNotExist) {
						return fmt.Errorf("removing config file: %w", err)
					}
				}

				continue
			}

			b, err := json.Marshal(cfg)
			if err != nil {
				return err
			}

			var dest bytes.Buffer
			if err := json.Indent(&dest, b, "", "  "); err != nil {
				return fmt.Errorf("formatting tools config: %w", err)
			}

			err = os.WriteFile(filepath.Join(s.storageFolder(), cfg.ID+".json"), dest.Bytes(), 0666)
			if err != nil {
				return err
			}
		}
		return nil
	}

	// we store in memory
	storage := map[string][]byte{}
	for _, cfg := range cfgs {
		b, err := json.Marshal(cfg)
		if err != nil {
			return err
		}

		var dest bytes.Buffer
		if err := json.Indent(&dest, b, "", "  "); err != nil {
			return fmt.Errorf("formatting tools config: %w", err)
		}
		storage[cfg.ID+".json"] = dest.Bytes()
	}

	s.ToolsConfigStorage = storage
	return nil
}

// addToProject will add the given tools to the existing project tools config and store them.
func (s *Service) addToProject(cfgs ...*ToolConfig) error {
	currentConfigs, err := s.load()
	if err != nil {
		return fmt.Errorf("loading tool configs: %w", err)
	}

	for _, toolConfig := range cfgs {
		if exists := currentConfigs.findByID(toolConfig.ID); exists != nil {
			return fmt.Errorf("tool config exists: %s", toolConfig.ID)
		}
		currentConfigs = append(currentConfigs, toolConfig)
	}

	if err := s.store(currentConfigs); err != nil {
		return fmt.Errorf("storing tool config to project: %w", err)
	}

	return nil
}

// updateToProject will replace the given tools in the stored config.
func (s *Service) updateToProject(cfgs ...*ToolConfig) error {
	currentConfigs, err := s.load()
	if err != nil {
		return fmt.Errorf("loading tools from config file: %w", err)
	}

	for _, updated := range cfgs {
		if currentConfigs.hasID(updated.ID) {
			for i := range currentConfigs {
				if currentConfigs[i].ID == updated.ID {
					currentConfigs[i] = updated
				}
			}
		} else {
			currentConfigs = append(currentConfigs, updated)
		}
	}

	if err := s.store(currentConfigs); err != nil {
		return fmt.Errorf("storing tools to project: %w", err)
	}

	return nil
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
	genTools, err := GenerateTools(ctx, s.Schema, s.Config)
	if err != nil {
		return nil, fmt.Errorf("generating tools from schema: %w", err)
	}

	tools := &toolsproto.Tools{
		Configs: genTools, // all tools, action based and/or flow based
	}

	// load existing configured tools
	configs, err := s.load()
	if err != nil {
		return nil, fmt.Errorf("loading tool configs from file: %w", err)
	}

	// if we have user added or configured tools...
	if len(configs) > 0 {
		// let's see get the user added tools
		addedIds := tools.DiffIDs(configs.getIDs())
		// for all the added ones, we generate a new tool and add it to our set
		for _, id := range addedIds {
			cfg := configs.findByID(id)
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
			if gen.GetActionConfig() != nil {
				gen.ActionConfig.Id = cfg.ID
			}

			tools.Configs = append(tools.Configs, gen)
		}

		// now we apply all the configs
		for _, cfg := range configs {
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
	if err := s.clear(); err != nil {
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

// getGeneratedTool will return the generated tool for the given action/flow name.
func (s *Service) getGeneratedTool(ctx context.Context, name string) (*toolsproto.Tool, error) {
	// if we don't have a schema, return nil
	if s.Schema == nil {
		return nil, nil
	}

	// generate tools
	genTools, err := GenerateTools(ctx, s.Schema, s.Config)
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

// GetFields returns the configured fields for this schema.
func (s *Service) GetFields(ctx context.Context) ([]*toolsproto.Field, error) {
	//TODO: implement
	return nil, nil
}

// ConfigureFields will take the given updated fields config and update the existing project config with it.
func (s *Service) ConfigureFields(ctx context.Context, updated []*toolsproto.Field) ([]*toolsproto.Field, error) {
	//TODO: implement
	return updated, nil
}
