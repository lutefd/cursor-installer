package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type CursorConfig struct {
	DisabledLanguages          []string `json:"cursor.cpp.disabledLanguages,omitempty"`
	EnablePartialAccepts       bool     `json:"cursor.cpp.enablePartialAccepts,omitempty"`
	UsePreviewBox              bool     `json:"cursor.terminal.usePreviewBox,omitempty"`
	RenderPillsInsteadOfBlocks bool     `json:"cursor.composer.renderPillsInsteadOfBlocks,omitempty"`
	CommandAllowlist           []string `json:"cursor.terminal.commandAllowlist,omitempty"`
	RequireApproval            bool     `json:"cursor.terminal.requireApproval,omitempty"`
	EnableYoloMode             bool     `json:"cursor.terminal.enableYoloMode,omitempty"`
}

func (i *Installer) ConfigureCursor() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err)
	}

	configDir := filepath.Join(homeDir, ".config", "Cursor", "User")
	settingsPath := filepath.Join(configDir, "settings.json")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	var settings map[string]interface{}
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to read settings file: %v", err)
	}

	if settings == nil {
		settings = make(map[string]interface{})
	}

	config := CursorConfig{
		DisabledLanguages:          []string{"scminput", "yaml"},
		EnablePartialAccepts:       true,
		UsePreviewBox:              true,
		RenderPillsInsteadOfBlocks: true,
		CommandAllowlist:           []string{"cd", "ls", "echo", "touch", "cp", "mv", "curl"},
		RequireApproval:            false,
		EnableYoloMode:             true,
	}

	configMap := make(map[string]interface{})
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	if err := json.Unmarshal(configData, &configMap); err != nil {
		return fmt.Errorf("failed to convert config to map: %v", err)
	}

	for k, v := range configMap {
		settings[k] = v
	}

	data, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %v", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings file: %v", err)
	}

	return nil
}
