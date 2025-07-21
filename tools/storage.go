package tools

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// storageFolder return the path to the tools config storage folder.
func (s *Service) storageFolder() string {
	if s.ProjectDir == nil {
		return ""
	}

	return filepath.Join(*s.ProjectDir, ToolsDir)
}

// hasFileStorage tells us if this service has been initialised with file storage.
func (s *Service) hasFileStorage() bool {
	return s.ProjectDir != nil
}

// initStorageFolder will create the folder in which user tool/field configs are stored, if applicable (service has file storage).
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
func (s *Service) loadFromFileStorage() (UserConfig, error) {
	if !s.hasFileStorage() {
		return UserConfig{}, fmt.Errorf("service does not have file storage enabled")
	}

	if err := s.initStorageFolder(); err != nil {
		return UserConfig{}, fmt.Errorf("initialising tools folder: %w", err)
	}

	configFiles, err := filepath.Glob(filepath.Join(s.storageFolder(), "*.json"))
	if err != nil {
		return UserConfig{}, err
	}
	userConfig := UserConfig{}

	for _, fName := range configFiles {
		fileBytes, err := os.ReadFile(fName)
		if err != nil {
			return UserConfig{}, err
		}

		if filepath.Base(fName) == FieldsFile {
			// read fields config
			var fCfg FieldConfigs
			if err := json.Unmarshal(fileBytes, &fCfg); err != nil {
				return UserConfig{}, err
			}
			userConfig.Fields = fCfg
		} else {
			// read tools config
			var tCfg ToolConfig
			if err := json.Unmarshal(fileBytes, &tCfg); err != nil {
				return UserConfig{}, err
			}
			userConfig.Tools = append(userConfig.Tools, &tCfg)
		}
	}

	return userConfig, nil
}

// load will read the tools and fields configurations from storage.
func (s *Service) load() (UserConfig, error) {
	// if we have file storage enabled, load from file
	if s.hasFileStorage() {
		return s.loadFromFileStorage()
	}

	// otherwise load from internal cache
	userConfig := UserConfig{}

	// read tools config
	for _, fileBytes := range s.ToolsConfigStorage {
		var cfg ToolConfig
		if err := json.Unmarshal(fileBytes, &cfg); err != nil {
			return UserConfig{}, err
		}
		userConfig.Tools = append(userConfig.Tools, &cfg)
	}

	// read fields config
	if len(s.FieldsConfigStorage) > 0 {
		if err := json.Unmarshal(s.FieldsConfigStorage, &userConfig.Fields); err != nil {
			return UserConfig{}, err
		}
	}

	return userConfig, nil
}

// clearTools will remove all the saved tool configs from the project.
func (s *Service) clearTools() error {
	if s.hasFileStorage() {
		files, err := os.ReadDir(s.storageFolder())
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.Name() == FieldsFile {
				continue
			}
			err := os.Remove(filepath.Join(s.storageFolder(), file.Name()))
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}

		return nil
	}

	s.ToolsConfigStorage = map[string][]byte{}

	return nil
}

// clearFields will remove all the saved fields configs from the project.
func (s *Service) clearFields() error {
	if s.hasFileStorage() {
		err := os.Remove(filepath.Join(s.storageFolder(), FieldsFile))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}

		return nil
	}

	s.FieldsConfigStorage = []byte{}
	return nil
}

// storeFields will save the given user configuration.
func (s *Service) storeFields(cfgs FieldConfigs) error {
	// if we have no changes, remove any existing changes
	if !cfgs.haveChanges() {
		return s.clearFields()
	}

	b, err := json.Marshal(cfgs.changed())
	if err != nil {
		return err
	}
	var dest bytes.Buffer
	if err := json.Indent(&dest, b, "", "  "); err != nil {
		return fmt.Errorf("formatting fields config: %w", err)
	}

	if s.hasFileStorage() {
		err = os.WriteFile(filepath.Join(s.storageFolder(), FieldsFile), dest.Bytes(), 0666)
		if err != nil {
			return err
		}
	}

	s.FieldsConfigStorage = dest.Bytes()

	return nil
}

// storeTools will save the given user configuration.
func (s *Service) storeTools(cfgs ToolConfigs) error {
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
	userConfig, err := s.load()
	if err != nil {
		return fmt.Errorf("loading tool configs: %w", err)
	}

	for _, toolConfig := range cfgs {
		if exists := userConfig.Tools.findByID(toolConfig.ID); exists != nil {
			return fmt.Errorf("tool config exists: %s", toolConfig.ID)
		}
		userConfig.Tools = append(userConfig.Tools, toolConfig)
	}

	if err := s.storeTools(userConfig.Tools); err != nil {
		return fmt.Errorf("storing tool config to project: %w", err)
	}

	return nil
}

// updateToProject will replace the given tools in the stored config.
func (s *Service) updateToProject(cfgs ...*ToolConfig) error {
	userConfig, err := s.load()
	if err != nil {
		return fmt.Errorf("loading tools from config file: %w", err)
	}

	for _, updated := range cfgs {
		if userConfig.Tools.hasID(updated.ID) {
			for i := range userConfig.Tools {
				if userConfig.Tools[i].ID == updated.ID {
					userConfig.Tools[i] = updated
				}
			}
		} else {
			userConfig.Tools = append(userConfig.Tools, updated)
		}
	}

	if err := s.storeTools(userConfig.Tools); err != nil {
		return fmt.Errorf("storing tools to project: %w", err)
	}

	return nil
}
