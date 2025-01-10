package app

import (
	"fmt"
	"io"
	"net/http"
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
	downloadOnly bool
	forceInstall bool
}

func NewInstaller(downloadOnly, forceInstall bool) *Installer {
	return &Installer{
		downloadOnly: downloadOnly,
		forceInstall: forceInstall,
	}
}

func (i *Installer) ensureInstallDir() error {
	cmd := exec.Command("sudo", "mkdir", "-p", installDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create install directory: %v", err)
	}
	return nil
}

func (i *Installer) CheckInstallation() error {
	metadata, err := i.readMetadata()
	if err != nil {
		return fmt.Errorf("failed to read installation metadata: %v", err)
	}

	if metadata == nil {
		_, err := os.Stat(filepath.Join(installDir, appImage))
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to check installation: %v", err)
		}
		if !i.forceInstall {
			return fmt.Errorf("cursor is already installed (legacy). Use --force to reinstall")
		}
		return nil
	}

	latestVersion, err := i.GetLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to check for updates: %v", err)
	}

	if latestVersion != metadata.Version || i.forceInstall {
		return nil
	}

	return fmt.Errorf("cursor version %s is already installed and up to date", metadata.Version)
}

func (i *Installer) DownloadCursor() error {
	resp, err := http.Get(cursorURL)
	if err != nil {
		return fmt.Errorf("failed to download Cursor: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(appImage)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to save download: %v", err)
	}

	return nil
}

func (i *Installer) MakeExecutable() error {
	cmd := exec.Command("chmod", "+x", appImage)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make file executable: %v", err)
	}
	return nil
}

func (i *Installer) MoveToOpt() error {
	if err := i.ensureInstallDir(); err != nil {
		return err
	}

	cmd := exec.Command("sudo", "mv", appImage, filepath.Join(installDir, appImage))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move file to %s: %v", installDir, err)
	}
	return nil
}

func (i *Installer) ExtractIcon() error {
	cmd := exec.Command(filepath.Join(installDir, appImage), "--appimage-extract")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract AppImage: %v", err)
	}

	iconPath := "squashfs-root/usr/share/icons/hicolor/512x512/apps/cursor.png"
	if _, err := os.Stat(iconPath); err != nil {
		return fmt.Errorf("icon not found in extracted contents: %v", err)
	}

	cmd = exec.Command("sudo", "mv", iconPath, filepath.Join(installDir, "cursor.png"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move icon: %v", err)
	}

	if err := os.RemoveAll("squashfs-root"); err != nil {
		return fmt.Errorf("failed to clean up extracted contents: %v", err)
	}

	return nil
}

func (i *Installer) CreateDesktopEntry() error {
	desktopEntry := fmt.Sprintf(`[Desktop Entry]
Name=Cursor
Exec=%s
Icon=%s
Type=Application
Categories=Development;
`, filepath.Join(installDir, appImage), filepath.Join(installDir, "cursor.png"))

	cmd := exec.Command("sudo", "bash", "-c", fmt.Sprintf("echo '%s' > /usr/share/applications/cursor.desktop", desktopEntry))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create desktop entry: %v", err)
	}
	return nil
}

func (i *Installer) CreateSymlink() error {
	cmd := exec.Command("sudo", "ln", "-sf", filepath.Join(installDir, appImage), "/usr/local/bin/cursor")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}
	return nil
}
