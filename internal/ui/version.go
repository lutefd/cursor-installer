package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lutefd/cursor-installer/internal/app"
)

type VersionDisplay struct {
	info *app.VersionInfo
	err  error
}

func NewVersionDisplay(info *app.VersionInfo, err error) *VersionDisplay {
	return &VersionDisplay{
		info: info,
		err:  err,
	}
}

func (v *VersionDisplay) View() string {
	if v.err != nil {
		return styleError.Render(fmt.Sprintf("Error getting version info: %v", v.err))
	}

	var s strings.Builder

	s.WriteString(versionHeaderStyle.Render("Cursor Versions") + "\n\n")

	header := []string{
		tableHeaderStyle.Render("Component"),
		tableHeaderStyle.Render("Version"),
		tableHeaderStyle.Render("Status"),
	}

	data := [][]string{
		{
			tableRowStyle.Render("Cursor"),
			tableValueStyle.Render(v.info.CursorVersion),
			tableRowStyle.Render(getInstallStatus(v.info.IsInstalled)),
		},
		{
			tableRowStyle.Render("Installer"),
			tableValueStyle.Render(v.info.InstallerVersion),
			tableRowStyle.Render("installed"),
		},
	}

	widths := make([]int, 3)
	for i, cell := range header {
		widths[i] = lipgloss.Width(cell)
	}

	for _, row := range data {
		for i, cell := range row {
			if w := lipgloss.Width(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}

	createBorder := func(left, mid, right, horizontal string) string {
		var border strings.Builder
		border.WriteString(tableBorderStyle.Render(left))
		for i, width := range widths {
			border.WriteString(strings.Repeat(tableBorderStyle.Render(horizontal), width+2))
			if i < len(widths)-1 {
				border.WriteString(tableBorderStyle.Render(mid))
			}
		}
		border.WriteString(tableBorderStyle.Render(right))
		return border.String()
	}

	s.WriteString(createBorder("┌", "┬", "┐", "─") + "\n")

	s.WriteString(tableBorderStyle.Render("│"))
	for i, cell := range header {
		padding := widths[i] - lipgloss.Width(cell)
		s.WriteString(" " + cell + strings.Repeat(" ", padding) + " " + tableBorderStyle.Render("│"))
	}
	s.WriteString("\n")

	s.WriteString(createBorder("├", "┼", "┤", "─") + "\n")

	for _, row := range data {
		s.WriteString(tableBorderStyle.Render("│"))
		for i, cell := range row {
			padding := widths[i] - lipgloss.Width(cell)
			s.WriteString(" " + cell + strings.Repeat(" ", padding) + " " + tableBorderStyle.Render("│"))
		}
		s.WriteString("\n")
	}

	s.WriteString(createBorder("└", "┴", "┘", "─") + "\n")

	return s.String()
}

func getInstallStatus(installed bool) string {
	if installed {
		return "installed"
	}
	return "not installed"
}
