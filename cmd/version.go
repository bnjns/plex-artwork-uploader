package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

const versionStringTemplate = `
version = %s
commit = %s
`

func SetVersionInfo(version string, commit string) {
	versionCmd.Version = fmt.Sprintf(strings.TrimSpace(versionStringTemplate), version, commit)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the current version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
