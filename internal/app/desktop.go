package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

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
	cmd = exec.Command("sudo", "-S", "cp", iconPath, targetPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy icon (sudo error): %v", err)
	}

	cmd = exec.Command("sudo", "-S", "chmod", "644", targetPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set icon permissions (sudo error): %v", err)
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

	cmd := exec.Command("sudo", "-S", "mv", tmpFile.Name(), "/usr/share/applications/cursor.desktop")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install desktop entry (sudo error): %v", err)
	}

	cmd = exec.Command("sudo", "-S", "chmod", "644", "/usr/share/applications/cursor.desktop")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set desktop entry permissions (sudo error): %v", err)
	}

	return nil
}

func (i *Installer) CreateSymlink() error {
	cmd := exec.Command("sudo", "-S", "ln", "-sf", filepath.Join(installDir, appImage), "/usr/local/bin/cursor")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create symlink (sudo error): %v", err)
	}
	return nil
}
