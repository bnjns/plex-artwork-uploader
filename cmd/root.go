package cmd

import (
	"github.com/bnjns/plex-artwork-uploader/cmd/manual"
	"github.com/bnjns/plex-artwork-uploader/cmd/mediux"
	"github.com/bnjns/plex-artwork-uploader/internal/flags"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "plex-artwork-uploader",
	Short:         "Uploads poster and background artwork to movies and shows in your Plex libraries.",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	_ = godotenv.Load()

	err := rootCmd.Execute()
	if err != nil {
		color.Red("\n%s", err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(manual.Cmd)
	rootCmd.AddCommand(mediux.Cmd)

	rootCmd.PersistentFlags().CountVarP(&flags.LogVerbosity, "verbose", "v", "Set the logging verbosity.")
	rootCmd.PersistentFlags().StringVar(&flags.PlexServerUrl, "server-url", "", "The base url of the plex server (eg, http://127.0.0.1:32400).")
	rootCmd.PersistentFlags().StringVar(&flags.PlexToken, "token", "", "The authentication token.")
}
