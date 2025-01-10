package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/lutefd/cursor-installer/internal/app"
)

type InstallationStep struct {
	name    string
	message string
	run     func() error
}

type model struct {
	spinner        spinner.Model
	currentStep    int
	completedSteps []bool
	err            error
	cancelled      bool
	completed      bool
	upToDate       bool
	currentVersion string
	installer      *app.Installer
	steps          []InstallationStep
	downloadOnly   bool
}

func checkInstallationWrapper(installer *app.Installer) func() error {
	return func() error {
		status := installer.CheckInstallation()
		return status.Error
	}
}

func NewModel(downloadOnly, forceInstall bool) model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))

	installer := app.NewInstaller(downloadOnly, forceInstall)

	var checkMessage string
	if downloadOnly && !forceInstall && !installer.CheckInstallation().AlreadyUpToDate {
		checkMessage = "Preparing to download..."
	} else {
		checkMessage = "Checking if Cursor is already installed..."
	}

	steps := []InstallationStep{
		{
			name:    "Check Installation",
			message: checkMessage,
			run:     checkInstallationWrapper(installer),
		},
	}

	info, err := installer.GetVersionInfo()
	if err == nil && info.IsInstalled && !forceInstall {
		var stepName, stepMessage string
		if downloadOnly {
			stepName = "Download"
			stepMessage = "Downloading latest version of Cursor..."
		} else {
			stepName = "Check Updates"
			stepMessage = "Checking for available updates..."
		}

		steps = append(steps, InstallationStep{
			name:    stepName,
			message: stepMessage,
			run: func() error {
				hasUpdate, err := installer.CheckForUpdates()
				if err != nil {
					return err
				}
				if !hasUpdate {
					return &upToDateError{version: info.CursorVersion}
				}
				return nil
			},
		})
	} else {
		steps = append(steps, InstallationStep{
			name:    "Download",
			message: "Downloading latest version of Cursor...",
			run: func() error {
				if err := installer.DownloadCursor(); err != nil {
					return err
				}
				return installer.MakeExecutable()
			},
		})
	}

	if !downloadOnly {
		steps = append(steps,
			InstallationStep{
				name:    "Install",
				message: "Moving Cursor to /opt directory...",
				run:     installer.MoveToOpt,
			},
			InstallationStep{
				name:    "Extract Icon",
				message: "Extracting application icon...",
				run:     installer.ExtractIcon,
			},
			InstallationStep{
				name:    "Desktop Entry",
				message: "Creating desktop entry...",
				run:     installer.CreateDesktopEntry,
			},
			InstallationStep{
				name:    "Create Symlink",
				message: "Creating symlink in /usr/local/bin...",
				run:     installer.CreateSymlink,
			},
			InstallationStep{
				name:    "Update Metadata",
				message: "Recording installation information...",
				run:     installer.UpdateMetadata,
			},
		)
	}

	return model{
		spinner:        s,
		installer:      installer,
		steps:          steps,
		completedSteps: make([]bool, len(steps)),
		downloadOnly:   downloadOnly,
	}
}
