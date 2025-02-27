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

const toolsFile = "tools.json"

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

// loadFromProject will read from the tools.json file the tools configuration.
// When a config fiel doesn't exist, a nil nil response will be returned
func (s *Service) loadFromProject() (*toolsproto.Tools, error) {
	path := filepath.Join(s.ProjectDir, toolsFile)

	if _, err := os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tools toolsproto.Tools
	if err := protojson.Unmarshal(fileBytes, &tools); err != nil {
		return nil, err
	}

	return &tools, nil
}

// storeToProject will save the given tools configuration to the tools.json file in the project.
func (s *Service) storeToProject(tools *toolsproto.Tools) error {
	path := filepath.Join(s.ProjectDir, toolsFile)

	opts := protojson.MarshalOptions{Indent: "  "}
	b, err := opts.Marshal(tools)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, b, 0666)
	if err != nil {
		return err
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

	updatedIds := []string{}
	for _, t := range tools {
		updatedIds = append(updatedIds, t.Id)
	}

	if !currentTools.HasIDs(updatedIds...) {
		return fmt.Errorf("one or more of the given tools do not exist")
	}

	for _, updated := range tools {
		for i, current := range currentTools.Tools {
			if current.Id == updated.Id {
				currentTools.Tools[i] = updated
			}
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

// GetTools will
// TODO: explain
func (s *Service) GetTools(ctx context.Context) (*toolsproto.Tools, error) {
	// if we don't have a schema, return nil
	if s.Schema == nil {
		return nil, nil
	}

	generated, err := GenerateTools(ctx, s.Schema, s.Config)
	if err != nil {
		return nil, fmt.Errorf("generating tools from schema: %w", err)
	}

	// load existing tool config
	existing, err := s.loadFromProject()
	if err != nil {
		return nil, fmt.Errorf("loading tools from config file: %w", err)
	}

	// no existing tools, we just save the generated ones
	if existing == nil {
		tools := toolsproto.Tools{
			Tools: generated,
		}

		if err := s.storeToProject(&tools); err != nil {
			return nil, fmt.Errorf("storing tools config: %w", err)
		}

		return &tools, nil
	}

	// we have existing tools, let's add newly generated ones
	if err := s.addToProject(existing.Diff(generated)...); err != nil {
		return nil, fmt.Errorf("adding new tool configs: %w", err)
	}

	// now we need to ensure that we sync new changes from the generated tools (e.g. add new inputs/responses/tool links, etc)
	if err := s.syncToProject(generated); err != nil {
		fmt.Println("Synced")
		return nil, fmt.Errorf("syncing tools configs: %w", err)
	}

	// reload from project
	return s.loadFromProject()
}

// ResetTools will replace the in-project tools config with freshly generated tools config from the current schema.
func (s *Service) ResetTools(ctx context.Context) (*toolsproto.Tools, error) {
	generated, err := GenerateTools(ctx, s.Schema, s.Config)
	if err != nil {
		return nil, err
	}

	tools := toolsproto.Tools{
		Tools: generated,
	}

	if err := s.storeToProject(&tools); err != nil {
		return nil, fmt.Errorf("storing tools to project: %w", err)
	}

	return &tools, nil
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
