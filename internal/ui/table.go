package ui

import "github.com/charmbracelet/lipgloss"

var (
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				PaddingRight(4)

	tableRowStyle = lipgloss.NewStyle().
			Foreground(textColor).
			PaddingRight(4)

	tableValueStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	tableBorderStyle = lipgloss.NewStyle().
				Foreground(primaryColor)

	versionHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor).
				MarginBottom(1).
				BorderStyle(lipgloss.RoundedBorder()).
				Padding(0, 1).
				Align(lipgloss.Center)
)
