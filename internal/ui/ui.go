package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lutefd/cursor-installer/internal/app"
)

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			MarginBottom(1)

	styleSuccess = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#59CD90"))

	styleError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	styleProgress = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4"))

	styleCompleted = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#59CD90")).
			SetString("✓")

	stylePending = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD93D")).
			SetString("•")
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
		{
			name:    "Download",
			message: "Downloading latest version of Cursor...",
			run:     installer.DownloadCursor,
		},
	}

	if !downloadOnly {
		steps = append(steps,
			InstallationStep{
				name:    "Make Executable",
				message: "Making the AppImage executable...",
				run:     installer.MakeExecutable,
			},
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
			return m, tea.Quit
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
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case upToDateMsg:
		m.upToDate = true
		m.currentVersion = msg.version
		return m, tea.Quit

	case doneMsg:
		m.completed = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() string {
	if m.cancelled {
		return styleError.Render("Installation cancelled by user.\n")
	}

	if m.upToDate {
		return styleSuccess.Render(fmt.Sprintf("✨ Cursor is already up to date (version %s)! ✨\n", m.currentVersion))
	}

	if m.completed {
		return styleSuccess.Render("✨ Cursor installation completed successfully! ✨\n")
	}

	if m.err != nil {
		return styleError.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	var s string
	s += styleTitle.Render("Cursor Installer") + "\n"
	s += styleProgress.Render(fmt.Sprintf("Progress: %d/%d", m.currentStep, len(m.steps))) + "\n"

	for i, step := range m.steps {
		var stepPrefix string
		if i < m.currentStep {
			stepPrefix = styleCompleted.String()
		} else if i == m.currentStep {
			stepPrefix = m.spinner.View()
		} else {
			stepPrefix = stylePending.String()
		}

		stepStyle := lipgloss.NewStyle()
		if i == m.currentStep {
			stepStyle = stepStyle.Foreground(lipgloss.Color("#4ECDC4"))
		} else if i < m.currentStep {
			stepStyle = stepStyle.Foreground(lipgloss.Color("#59CD90"))
		}

		s += fmt.Sprintf("%s %s", stepPrefix, stepStyle.Render(step.name))
		if i == m.currentStep {
			s += fmt.Sprintf(": %s", step.message)
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
