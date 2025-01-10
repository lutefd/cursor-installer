package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

func (i *Installer) ensureInstallDir() error {
	cmd := exec.Command("sudo", "mkdir", "-p", installDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create install directory: %v", err)
	}

	cmd = exec.Command("sudo", "chmod", "755", installDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set permissions on install directory: %v", err)
	}

	return nil
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

	latestVersion, err := i.GetLatestVersion()
	if err != nil {
		return &InstallationStatus{Error: fmt.Errorf("failed to check for updates: %v", err)}
	}

	if latestVersion != metadata.Version || i.forceInstall {
		return &InstallationStatus{}
	}

	return &InstallationStatus{
		AlreadyUpToDate: true,
		CurrentVersion:  metadata.Version,
	}
}

func (i *Installer) DownloadCursor() error {
	resp, err := http.Get(cursorURL)
	if err != nil {
		return fmt.Errorf("failed to download Cursor: %v", err)
	}
	defer resp.Body.Close()

	contentDisposition := resp.Header.Get("Content-Disposition")
	var originalFilename string
	if strings.Contains(contentDisposition, "filename=") {
		originalFilename = strings.Split(contentDisposition, "filename=")[1]
		originalFilename = strings.Trim(originalFilename, "\"")
	}

	if originalFilename != "" {
		re := regexp.MustCompile(`cursor-(.+?)(?:x86_64)?\.AppImage`)
		matches := re.FindStringSubmatch(originalFilename)
		if len(matches) > 1 {
			i.version = matches[1]
		}
	}

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
	cmd := exec.Command("sudo", "chmod", "+x", appImage)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make file executable: %v", err)
	}
	return nil
}

func (i *Installer) MoveToOpt() error {
	if err := i.ensureInstallDir(); err != nil {
		return err
	}

	targetPath := filepath.Join(installDir, appImage)
	cmd := exec.Command("sudo", "mv", appImage, targetPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move file to %s: %v", installDir, err)
	}

	cmd = exec.Command("sudo", "chmod", "755", targetPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set permissions: %v", err)
	}

	return nil
}

func (i *Installer) ExtractIcon() error {
	tempDir, err := os.MkdirTemp("", "cursor-icon")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		return fmt.Errorf("failed to change to temp directory: %v", err)
	}
	defer os.Chdir(currentDir)

	cmd := exec.Command(filepath.Join(installDir, appImage), "--appimage-extract")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to extract AppImage: %v", err)
	}

	iconPath := "squashfs-root/usr/share/icons/hicolor/512x512/apps/cursor.png"
	if _, err := os.Stat(iconPath); err != nil {
		return fmt.Errorf("icon not found in extracted contents: %v", err)
	}

	targetPath := filepath.Join(installDir, "cursor.png")
	cmd = exec.Command("sudo", "cp", iconPath, targetPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy icon: %v", err)
	}

	cmd = exec.Command("sudo", "chmod", "644", targetPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set icon permissions: %v", err)
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

	tmpFile, err := os.CreateTemp("", "cursor-*.desktop")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(desktopEntry); err != nil {
		return fmt.Errorf("failed to write desktop entry: %v", err)
	}
	tmpFile.Close()

	cmd := exec.Command("sudo", "mv", tmpFile.Name(), "/usr/share/applications/cursor.desktop")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install desktop entry: %v", err)
	}

	cmd = exec.Command("sudo", "chmod", "644", "/usr/share/applications/cursor.desktop")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set desktop entry permissions: %v", err)
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
