package ui

import "github.com/charmbracelet/lipgloss"

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

	styleFilePath = lipgloss.NewStyle().
			Foreground(primaryColor).
			Italic(true).
			Bold(true)

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
