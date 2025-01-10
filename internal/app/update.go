package app

import (
	"fmt"
	"os"
)

func (i *Installer) CheckForUpdates() (bool, error) {
	metadata, err := i.readMetadata()
	if err != nil {
		return false, fmt.Errorf("failed to read metadata: %v", err)
	}

	if err := i.DownloadCursor(); err != nil {
		return false, fmt.Errorf("failed to check for updates: %v", err)
	}

	latestVersion := i.version
	if latestVersion == "" {
		os.Remove(appImage)
		return false, fmt.Errorf("failed to determine latest version")
	}

	if metadata == nil {
		os.Remove(appImage)
		return false, fmt.Errorf("no metadata found, cannot compare versions")
	}

	needsUpdate := latestVersion != metadata.Version
	if !needsUpdate {
		os.Remove(appImage)
	} else {
		if err := i.MakeExecutable(); err != nil {
			os.Remove(appImage)
			return false, fmt.Errorf("failed to make file executable: %v", err)
		}
	}

	return needsUpdate, nil
}
