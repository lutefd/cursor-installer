package app

import (
	"fmt"
	"os"
	"path/filepath"
)

const InstallerVersion = "0.3.0"

type VersionInfo struct {
	CursorVersion    string
	InstallerVersion string
	IsInstalled      bool
}

func (i *Installer) GetVersionInfo() (*VersionInfo, error) {
	info := &VersionInfo{
		InstallerVersion: InstallerVersion,
	}

	cursorPath := filepath.Join(installDir, appImage)
	if _, err := os.Stat(cursorPath); err != nil {
		if os.IsNotExist(err) {
			info.IsInstalled = false
			return info, nil
		}
		return nil, fmt.Errorf("failed to check cursor installation: %v", err)
	}

	metadata, err := i.readMetadata()
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %v", err)
	}

	if metadata != nil {
		info.CursorVersion = metadata.Version
		info.IsInstalled = true
	} else {
		info.CursorVersion = "unknown"
		info.IsInstalled = true
	}

	return info, nil
}
