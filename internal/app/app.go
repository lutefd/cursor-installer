package app

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	cursorURL  = "https://downloader.cursor.sh/linux/appImage/x64"
	appImage   = "Cursor.AppImage"
	installDir = "/opt/cursor"
)

type Installer struct {
	downloadOnly bool
	forceInstall bool
	version      string
}

type InstallationStatus struct {
	AlreadyUpToDate bool
	CurrentVersion  string
	Error           error
}

func NewInstaller(downloadOnly, forceInstall bool) *Installer {
	return &Installer{
		downloadOnly: downloadOnly,
		forceInstall: forceInstall,
	}
}

func (i *Installer) CheckInstallation() *InstallationStatus {
	metadata, err := i.readMetadata()
	if err != nil {
		return &InstallationStatus{Error: fmt.Errorf("failed to read installation metadata: %v", err)}
	}

	if metadata == nil {
		_, err := os.Stat(filepath.Join(installDir, appImage))
		if os.IsNotExist(err) {
			return &InstallationStatus{}
		}
		if err != nil {
			return &InstallationStatus{Error: fmt.Errorf("failed to check installation: %v", err)}
		}
		if !i.forceInstall {
			return &InstallationStatus{Error: fmt.Errorf("cursor is already installed (legacy). Use --force to reinstall")}
		}
		return &InstallationStatus{}
	}

	if i.forceInstall {
		return &InstallationStatus{}
	}

	return &InstallationStatus{
		AlreadyUpToDate: false,
		CurrentVersion:  metadata.Version,
	}
}
