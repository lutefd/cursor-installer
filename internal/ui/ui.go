package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lutefd/cursor-installer/internal/app"
)

var (
	primaryColor   = lipgloss.Color("#8BE9FD")
	successColor   = lipgloss.Color("#50FA7B")
	errorColor     = lipgloss.Color("#FF5555")
	warningColor   = lipgloss.Color("#FFB86C")
	secondaryColor = lipgloss.Color("#BD93F9")
	textColor      = lipgloss.Color("#F8F8F2")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1).
			BorderStyle(lipgloss.RoundedBorder()).
			Padding(0, 1).
			Align(lipgloss.Center)

	styleSuccess = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			PaddingLeft(2)

	styleError = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			PaddingLeft(2)

	styleProgress = lipgloss.NewStyle().
			Foreground(secondaryColor).
			PaddingLeft(2)

	styleCompleted = lipgloss.NewStyle().
			Foreground(successColor).
			SetString("✓").
			PaddingRight(1)

	stylePending = lipgloss.NewStyle().
			Foreground(warningColor).
			SetString("•").
			PaddingRight(1)

	styleCurrentStep = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	styleCompletedStep = lipgloss.NewStyle().
				Foreground(textColor).
				Faint(true)

	stylePendingStep = lipgloss.NewStyle().
				Foreground(textColor).
				Faint(true)

	styleStepMessage = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Italic(true)
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

	steps := []InstallationStep{
		{
			name:    "Check Installation",
			message: "Checking if Cursor is already installed...",
			run:     checkInstallationWrapper(installer),
		},
	}

	info, err := installer.GetVersionInfo()
	if err == nil && info.IsInstalled && !forceInstall {
		steps = append(steps, InstallationStep{
			name:    "Check Updates",
			message: "Checking for available updates...",
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
	}
}

// Custom error type for up-to-date case
type upToDateError struct {
	version string
}

func (e *upToDateError) Error() string {
	return "already up to date"
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runNextStep(),
	)
}

func (m model) runNextStep() tea.Cmd {
	return func() tea.Msg {
		if m.currentStep >= len(m.steps) {
			return doneMsg{}
		}

		step := m.steps[m.currentStep]

		if m.currentStep == 0 {
			status := m.installer.CheckInstallation()
			if status.Error != nil {
				return errMsg(status.Error)
			}
			if status.AlreadyUpToDate {
				return upToDateMsg{version: status.CurrentVersion}
			}
		} else {
			if err := step.run(); err != nil {
				if upToDateErr, ok := err.(*upToDateError); ok {
					return upToDateMsg{version: upToDateErr.version}
				}
				return errMsg(err)
			}
		}

		return stepCompleteMsg{
			stepName: step.name,
			nextStep: m.currentStep + 1,
		}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.cancelled = true
			return m, tea.Sequence(
				tea.Println(styleError.Render("Installation cancelled by user")),
				tea.Quit,
			)
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case stepCompleteMsg:
		m.completedSteps[m.currentStep] = true
		m.currentStep = msg.nextStep
		if m.currentStep < len(m.steps) {
			return m, m.runNextStep()
		}
		m.completed = true
		return m, tea.Sequence(
			tea.Println(styleSuccess.Render("✨ Cursor installation completed successfully! ✨")),
			tea.Quit,
		)

	case errMsg:
		m.err = msg
		return m, tea.Sequence(
			tea.Println(styleError.Render(fmt.Sprintf("Error: %v", m.err))),
			tea.Quit,
		)

	case upToDateMsg:
		m.upToDate = true
		m.currentVersion = msg.version
		return m, tea.Sequence(
			tea.Println(styleSuccess.Render("✨ Cursor "+m.currentVersion+" is already installed and up to date! ✨")),
			tea.Quit,
		)

	case doneMsg:
		m.completed = true
		return m, tea.Sequence(
			tea.Println(styleSuccess.Render("✨ Cursor installation completed successfully! ✨")),
			tea.Quit,
		)
	}

	return m, nil
}

func (m model) View() string {
	if m.cancelled {
		return styleError.Render("✗ Installation cancelled by user")
	}

	if m.upToDate {
		return styleSuccess.Render(fmt.Sprintf("✨ Cursor is already up to date (version %s)!", m.currentVersion))
	}

	if m.completed {
		return styleSuccess.Render("✨ Cursor installation completed successfully!")
	}

	if m.err != nil {
		return styleError.Render(fmt.Sprintf("✗ Error: %v", m.err))
	}

	var s string

	s += styleTitle.Render("Cursor Installer") + "\n\n"

	progress := fmt.Sprintf("Step %d of %d", m.currentStep+1, len(m.steps))
	s += styleProgress.Render(progress) + "\n\n"

	for i, step := range m.steps {
		var stepPrefix string
		var stepStyle lipgloss.Style

		if i < m.currentStep {
			stepPrefix = styleCompleted.String()
			stepStyle = styleCompletedStep
		} else if i == m.currentStep {
			stepPrefix = m.spinner.View()
			stepStyle = styleCurrentStep
		} else {
			stepPrefix = stylePending.String()
			stepStyle = stylePendingStep
		}

		s += fmt.Sprintf("%s%s", stepPrefix, stepStyle.Render(step.name))

		if i == m.currentStep {
			s += styleStepMessage.Render(fmt.Sprintf(": %s", step.message))
		}

		s += "\n"
	}

	return s
}

type stepCompleteMsg struct {
	stepName string
	nextStep int
}

type errMsg error
type doneMsg struct{}
type upToDateMsg struct {
	version string
}

func GetLongDescription() string {
	return `Cursor Installer is a tool to download and install the Cursor editor.
It handles downloading the latest version, setting up desktop integration,
and creating necessary symlinks.`
}
