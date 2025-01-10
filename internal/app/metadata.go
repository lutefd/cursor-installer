package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type CursorMetadata struct {
	Version        string    `json:"version"`
	InstallDate    time.Time `json:"install_date"`
	LastUpdateDate time.Time `json:"last_update_date"`
	InstallPath    string    `json:"install_path"`
}

const metadataPath = "/opt/cursor/metadata.json"

func (i *Installer) GetLatestVersion() (string, error) {
	resp, err := http.Head(cursorURL)
	if err != nil {
		return "", fmt.Errorf("failed to get latest version: %v", err)
	}
	defer resp.Body.Close()

	version := resp.Header.Get("X-Version")
	if version == "" {
		version = "latest"
	}

	return version, nil
}

func (i *Installer) readMetadata() (*CursorMetadata, error) {
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read metadata: %v", err)
	}

	var metadata CursorMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %v", err)
	}

	return &metadata, nil
}

func (i *Installer) writeMetadata(metadata *CursorMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %v", err)
	}

	return nil
}

func (i *Installer) UpdateMetadata() error {
	if err := i.ensureInstallDir(); err != nil {
		return err
	}

	latestVersion, err := i.GetLatestVersion()
	if err != nil {
		return err
	}

	metadata := &CursorMetadata{
		Version:        latestVersion,
		InstallPath:    filepath.Join(installDir, appImage),
		LastUpdateDate: time.Now(),
	}

	existingMetadata, err := i.readMetadata()
	if err != nil {
		return err
	}

	if existingMetadata != nil {
		metadata.InstallDate = existingMetadata.InstallDate
	} else {
		metadata.InstallDate = time.Now()
	}

	return i.writeMetadata(metadata)
}
