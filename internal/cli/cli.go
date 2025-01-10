package cli

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lutefd/cursor-installer/internal/app"
	"github.com/lutefd/cursor-installer/internal/ui"
	"github.com/spf13/cobra"
)

var (
	downloadOnly bool
	forceInstall bool
	showVersion  bool
)

func Execute() error {
	var rootCmd = &cobra.Command{
		Use:   "cursor-installer",
		Short: "Install Cursor Editor",
		Long:  ui.GetLongDescription(),
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				installer := app.NewInstaller(false, false)
				info, err := installer.GetVersionInfo()
				display := ui.NewVersionDisplay(info, err)
				fmt.Println(display.View())
				return nil
			}

			model := ui.NewModel(downloadOnly, forceInstall)
			program := tea.NewProgram(model)

			if _, err := program.Run(); err != nil {
				return fmt.Errorf("installation failed: %v", err)
			}
			return nil
		},
	}

	rootCmd.Flags().BoolVarP(&downloadOnly, "download-only", "d", false, "Only download Cursor without installing")
	rootCmd.Flags().BoolVarP(&forceInstall, "force", "f", false, "Force installation even if Cursor is already installed")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Display version information")

	return rootCmd.Execute()
}
