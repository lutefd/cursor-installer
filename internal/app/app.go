package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
)

const (
	cursorURL = "https://downloader.cursor.sh/linux/appImage/x64"
	appImage  = "cursor.AppImage"
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

func (i *Installer) CheckInstallation() error {
	_, err := os.Stat("/opt/cursor.AppImage")
	if err == nil && !i.forceInstall {
		return fmt.Errorf("cursor is already installed. Use --force to reinstall")
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to check installation: %v", err)
	}
	return nil
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
	cmd := exec.Command("sudo", "mv", appImage, "/opt/cursor.AppImage")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move file to /opt/: %v", err)
	}
	return nil
}

func (i *Installer) ExtractIcon() error {
	cmd := exec.Command("/opt/cursor.AppImage", "--appimage-extract")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract AppImage: %v", err)
	}

	iconPath := "squashfs-root/usr/share/icons/hicolor/512x512/apps/cursor.png"
	if _, err := os.Stat(iconPath); err != nil {
		return fmt.Errorf("icon not found in extracted contents: %v", err)
	}

	cmd = exec.Command("sudo", "mv", iconPath, "/opt/cursor.png")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move icon: %v", err)
	}

	if err := os.RemoveAll("squashfs-root"); err != nil {
		return fmt.Errorf("failed to clean up extracted contents: %v", err)
	}

	return nil
}

func (i *Installer) CreateDesktopEntry() error {
	desktopEntry := `[Desktop Entry]
Name=Cursor
Exec=/opt/cursor.AppImage
Icon=/opt/cursor.png
Type=Application
Categories=Development;
`
	cmd := exec.Command("sudo", "bash", "-c", fmt.Sprintf("echo '%s' > /usr/share/applications/cursor.desktop", desktopEntry))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create desktop entry: %v", err)
	}
	return nil
}

func (i *Installer) CreateSymlink() error {
	cmd := exec.Command("sudo", "ln", "-sf", "/opt/cursor.AppImage", "/usr/local/bin/cursor")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create symlink: %v", err)
	}
	return nil
}
