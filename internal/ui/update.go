package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runNextStep(),
	)
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
		successMsg := "✨ Cursor installation completed successfully! ✨"
		if m.downloadOnly {
			pwd, _ := os.Getwd()
			filePath := filepath.Join(pwd, appImage)
			successMsg = fmt.Sprintf("✨ Cursor downloaded successfully to %s ✨", styleFilePath.Render(filePath))
		}
		return m, tea.Sequence(
			tea.Println(styleSuccess.Render(successMsg)),
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
		successMsg := "✨ Cursor installation completed successfully! ✨"
		if m.downloadOnly {
			pwd, _ := os.Getwd()
			filePath := filepath.Join(pwd, appImage)
			successMsg = fmt.Sprintf("✨ Cursor downloaded successfully to %s ✨", styleFilePath.Render(filePath))
		}
		return m, tea.Sequence(
			tea.Println(styleSuccess.Render(successMsg)),
			tea.Quit,
		)
	}

	return m, nil
}
