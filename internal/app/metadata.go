package app

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const metadataPath = "/opt/cursor/metadata.json"

type CursorMetadata struct {
	Version        string    `json:"version"`
	InstallDate    time.Time `json:"install_date"`
	LastUpdateDate time.Time `json:"last_update_date"`
	InstallPath    string    `json:"install_path"`
}

func (i *Installer) GetLatestVersion() (string, error) {
	if i.version != "" {
		return i.version, nil
	}
	return "unknown", nil
}

func (i *Installer) readMetadata() (*CursorMetadata, error) {
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, nil
	}

	tmpFile, err := os.CreateTemp("", "cursor-metadata-read-*.json")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	cmd := exec.Command("sudo", "-S", "sh", "-c", fmt.Sprintf("cp %s %s && chmod 644 %s", metadataPath, tmpPath, tmpPath))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to copy and set permissions on metadata file (sudo error): %v", err)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %v", err)
	}

	var metadata CursorMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %v", err)
	}

	return &metadata, nil
}

func (i *Installer) writeMetadata(metadata *CursorMetadata) error {
	tmpFile, err := os.CreateTemp("", "cursor-metadata-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temporary metadata file: %v", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata: %v", err)
	}
	tmpFile.Close()

	cmd := exec.Command("sudo", "-S", "sh", "-c", fmt.Sprintf("mv %s %s && chmod 644 %s", tmpPath, metadataPath, metadataPath))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install and set permissions on metadata file (sudo error): %v", err)
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
