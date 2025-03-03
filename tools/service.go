package tools

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	gonanoid "github.com/matoous/go-nanoid/v2"

	"github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/proto"
	toolsproto "github.com/teamkeel/keel/tools/proto"
	"google.golang.org/protobuf/encoding/protojson"
)

const toolsDir = "tools"

const (
	// Alphabet for unique nanoids to be used for tool ids
	nanoidABC = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Size of unique nanoids to be used for tool ids
	nanoidSize = 5
)

type Service struct {
	Schema     *proto.Schema
	Config     *config.ProjectConfig
	ProjectDir string
}

type ServiceParams struct {
	ProjectDir string
}

type ServiceOpt func(s *Service)

func NewService(params ServiceParams, opts ...ServiceOpt) (*Service, error) {
	svc := &Service{
		ProjectDir: params.ProjectDir,
	}

	for _, o := range opts {
		o(svc)
	}

	if err := svc.validate(); err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *Service) validate() error {
	if s.ProjectDir == "" {
		return errors.New("tools service: project dir required")
	}

	return nil
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

func (s *Service) initToolsFolder() error {
	path := filepath.Join(s.ProjectDir, toolsDir)
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("initialising tools dir: %w", err)
		}
	}
	if err != nil {
		return fmt.Errorf("initialising tools dir: %w", err)
	}
	return nil
}

// loadFromProject will read from the tools.json file the tools configuration.
// When a config fiel doesn't exist, a nil nil response will be returned
func (s *Service) loadFromProject() (*toolsproto.Tools, error) {
	configFiles, err := filepath.Glob(filepath.Join(s.ProjectDir, toolsDir, "*.json"))
	if err != nil {
		return nil, err
	}

	tools := toolsproto.Tools{}

	for _, fName := range configFiles {
		fileBytes, err := os.ReadFile(fName)
		if err != nil {
			return nil, err
		}
		var cfg toolsproto.ActionConfig
		if err := protojson.Unmarshal(fileBytes, &cfg); err != nil {
			return nil, err
		}
		tools.Tools = append(tools.Tools, &cfg)
	}

	return &tools, nil
}

// clearProject will remove all the saved tool configs from the project.
func (s *Service) clearProject() error {
	path := filepath.Join(s.ProjectDir, toolsDir)

	err := os.RemoveAll(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	return nil
}

// storeToProject will save the given tools configuration to the tools.json file in the project.
func (s *Service) storeToProject(tools *toolsproto.Tools) error {
	if err := s.initToolsFolder(); err != nil {
		return fmt.Errorf("initialising tools folder: %w", err)
	}

	opts := protojson.MarshalOptions{Indent: "  "}
	for _, cfg := range tools.Tools {
		b, err := opts.Marshal(cfg)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(s.ProjectDir, toolsDir, cfg.Id+".json"), b, 0666)
		if err != nil {
			return err
		}
	}

	return nil
}

// addToProject will add the given tools to the existing project tools config and store them.
func (s *Service) addToProject(tools ...*toolsproto.ActionConfig) error {
	currentTools, err := s.loadFromProject()
	if err != nil {
		return fmt.Errorf("loading tools from config file: %w", err)
	}

	for _, toolConfig := range tools {
		if exists := currentTools.FindByID(toolConfig.Id); exists != nil {
			return fmt.Errorf("tool already exists: %s", toolConfig.Id)
		}
		currentTools.Tools = append(currentTools.Tools, toolConfig)
	}

	if err := s.storeToProject(currentTools); err != nil {
		return fmt.Errorf("storing tools to project: %w", err)
	}

	return nil
}

// updateToProject will replace the given tools in the stored config.
func (s *Service) updateToProject(tools ...*toolsproto.ActionConfig) error {
	currentTools, err := s.loadFromProject()
	if err != nil {
		return fmt.Errorf("loading tools from config file: %w", err)
	}

	for _, updated := range tools {
		if currentTools.HasIDs(updated.Id) {
			for i := range currentTools.Tools {
				if currentTools.Tools[i].Id == updated.Id {
					currentTools.Tools[i] = updated
				}
			}
		} else {
			currentTools.Tools = append(currentTools.Tools, updated)
		}
	}

	if err := s.storeToProject(currentTools); err != nil {
		return fmt.Errorf("storing tools to project: %w", err)
	}

	return nil
}

// syncToProject will ensure that the existing tool config is synced with the generated configs. Changes will be
// persisted in the project tool config. These changes include:
// - adding new inputs
// - adding new responses
// - adding new tool links
// TODO: more syncing
func (s *Service) syncToProject(generated []*toolsproto.ActionConfig) error {
	currentTools, err := s.loadFromProject()
	if err != nil {
		return fmt.Errorf("loading tools from config file: %w", err)
	}

	// if we don't have any configured tools, return
	if !currentTools.HasTools() {
		return nil
	}

	for _, t := range currentTools.Tools {
		// for each tool, we find the generated one for the same action
		gen := toolsproto.FindByAction(generated, t.ActionName)
		if gen == nil {
			// this tool is no longer valid as the underlying action was removed, skip
			continue
		}

		syncTool(t, gen)
	}

	if err := s.storeToProject(currentTools); err != nil {
		return fmt.Errorf("storing tools to project: %w", err)
	}

	return nil
}

// DuplicateTool will take the given tool and duplicate it with a new ID, and then store the changes to files.
func (s *Service) DuplicateTool(ctx context.Context, toolID string) (*toolsproto.ActionConfig, error) {
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

	duplicate.Id += "-" + uid
	duplicate.Name += " (copy)"

	if err := s.addToProject(duplicate); err != nil {
		return nil, fmt.Errorf("duplicating tool: %w", err)
	}

	return duplicate, nil
}

// GetTools generates tools based on the schema, reads the configured tools from the project and applies them to the
// generated ones, returning a complete list of tool configs
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
		Tools: genTools,
	}

	// now we need to ensure that we sync new changes from the generated tools (e.g. add new inputs/responses/tool links, etc)
	if err := s.syncToProject(genTools); err != nil {
		return nil, fmt.Errorf("syncing tools configs: %w", err)
	}

	// load existing configured tools
	existing, err := s.loadFromProject()
	if err != nil {
		return nil, fmt.Errorf("loading tools from config file: %w", err)
	}

	// if we have user added or configured tools...
	if existing.HasTools() {
		// append the user added ones
		tools.Tools = append(tools.Tools, tools.Diff(existing.Tools)...)

		for _, id := range tools.IntersectIDs(existing) {
			configured := existing.FindByID(id)

			for i := range tools.Tools {
				if tools.Tools[i].Id == id {
					tools.Tools[i] = configured
				}
			}
		}
	}

	return tools, nil
}

// ResetTools will remove all the saved tool configs and return the schema generated tools.
func (s *Service) ResetTools(ctx context.Context) (*toolsproto.Tools, error) {
	if err := s.clearProject(); err != nil {
		return nil, fmt.Errorf("removing saved tools configuration: %w", err)
	}

	return s.GetTools(ctx)
}

// ConfigureTool will take the given updated tool config and update the existing project config with it
func (s *Service) ConfigureTool(ctx context.Context, updated *toolsproto.ActionConfig) (*toolsproto.ActionConfig, error) {
	// update the tool saved in project
	err := s.updateToProject(updated)
	if err != nil {
		return nil, fmt.Errorf("updating tool %s: %w", updated.GetId(), err)
	}

	// read them and return the fresh copy
	tools, err := s.GetTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving tools: %w", err)
	}

	return tools.FindByID(updated.Id), nil
}
