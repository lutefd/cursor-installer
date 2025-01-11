package app

import (
	"fmt"
	"os"
	"os/exec"
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

func (i *Installer) CheckSudoAccess() error {
	cmd := exec.Command("sudo", "-n", "true")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("this installer requires sudo privileges. Please ensure you have sudo access and try again. You can run `sudo usermod -aG sudo <user>` to add your user to the sudoers group")
	}
	return nil
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
