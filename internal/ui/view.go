package ui

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.cancelled {
		return styleError.Render("✗ Installation cancelled by user")
	}

	if m.upToDate {
		return styleSuccess.Render(fmt.Sprintf("✨ Cursor is already up to date (version %s)!", m.currentVersion))
	}

	if m.completed {
		if m.downloadOnly {
			pwd, _ := os.Getwd()
			filePath := filepath.Join(pwd, appImage)
			return styleSuccess.Render(fmt.Sprintf("✨ Cursor downloaded successfully to %s ✨", styleFilePath.Render(filePath)))
		}
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
