package cmd

import (
	"github.com/chain710/manga/internal/version"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show command version",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(version.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
