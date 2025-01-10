package ui

import tea "github.com/charmbracelet/bubbletea"

func (m model) runNextStep() tea.Cmd {
	return func() tea.Msg {
		if m.currentStep >= len(m.steps) {
			return doneMsg{}
		}

		step := m.steps[m.currentStep]

		if m.currentStep == 0 && !m.downloadOnly {
			status := m.installer.CheckInstallation()
			if status.Error != nil {
				return errMsg(status.Error)
			}
			if status.AlreadyUpToDate {
				return upToDateMsg{version: status.CurrentVersion}
			}
		}

		if err := step.run(); err != nil {
			if upToDateErr, ok := err.(*upToDateError); ok {
				return upToDateMsg{version: upToDateErr.version}
			}
			return errMsg(err)
		}

		return stepCompleteMsg{
			stepName: step.name,
			nextStep: m.currentStep + 1,
		}
	}
}
