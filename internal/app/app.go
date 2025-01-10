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
	downloadOnly      bool
	forceInstall      bool
	configureSettings bool
	version           string
}

type InstallationStatus struct {
	AlreadyUpToDate bool
	CurrentVersion  string
	Error           error
}

func NewInstaller(downloadOnly, forceInstall bool, configureSettings bool) *Installer {
	return &Installer{
		downloadOnly:      downloadOnly,
		forceInstall:      forceInstall,
		configureSettings: configureSettings,
	}
}

func (i *Installer) CheckInstallation() *InstallationStatus {
	metadata, err := i.readMetadata()
	if err != nil {
		return &InstallationStatus{Error: fmt.Errorf("failed to read installation metadata: %v", err)}
	}

	_, err = os.Stat(filepath.Join(installDir, appImage))
	if os.IsNotExist(err) {
		return &InstallationStatus{}
	}
	if err != nil {
		return &InstallationStatus{Error: fmt.Errorf("failed to check installation: %v", err)}
	}

	if i.forceInstall {
		return &InstallationStatus{}
	}

	if metadata != nil {
		return &InstallationStatus{
			AlreadyUpToDate: false,
			CurrentVersion:  metadata.Version,
		}
	}

	return &InstallationStatus{
		AlreadyUpToDate: false,
		CurrentVersion:  "unknown",
	}
}
