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

func (i *Installer) ensureInstallDir() error {
	cmd := exec.Command("sudo", "-S", "mkdir", "-p", installDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create install directory (sudo error): %v", err)
	}

	cmd = exec.Command("sudo", "-S", "chmod", "755", installDir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set permissions on install directory (sudo error): %v", err)
	}

	return nil
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
		re, err := regexp.Compile(`cursor-(.+?)(?:x86_64)?\.AppImage`)
		if err != nil {
			return fmt.Errorf("failed to compile regex: %v", err)
		}
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
	cmd := exec.Command("sudo", "-S", "chmod", "+x", appImage)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to make file executable (sudo error): %v", err)
	}
	return nil
}

func (i *Installer) MoveToOpt() error {
	if err := i.ensureInstallDir(); err != nil {
		return err
	}

	targetPath := filepath.Join(installDir, appImage)
	cmd := exec.Command("sudo", "-S", "mv", appImage, targetPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move file to %s (sudo error): %v", installDir, err)
	}

	cmd = exec.Command("sudo", "-S", "chmod", "755", targetPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set permissions (sudo error): %v", err)
	}

	return nil
}
