package mediux

import (
	"fmt"
	"github.com/bnjns/plex-artwork-uploader/internal/config"
	"github.com/bnjns/plex-artwork-uploader/internal/logger"
	"github.com/bnjns/plex-artwork-uploader/internal/mediux"
	"github.com/bnjns/plex-artwork-uploader/internal/plex"
	"github.com/bnjns/plex-artwork-uploader/internal/types"
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:   "file [path]",
	Short: "Import artwork using a MediUX YAML file",
	Long: `Use this to import poster and background artwork using a MediUX-style YAML file, although the artwork doesn't
need to actually be from MediUX.com.

The YAML file must be organised by Library and follow the structure:

'<library title>':
  '<movie id>':
    [url_poster: <url>]
    [url_background: <url>]
  '<show id>':
    [url_poster: <url>]
    [url_background: <url>]
    seasons:
      <season num>:
        [url_poster: <url>]
        [url_background: <url>]
        episodes:
          <episode num>:
            [url_poster: <url>]
            [url_background: <url>]

This assumes that movies are specified using an ID from TMDB and shows from TVDB.
`,
	Args: cobra.MatchAll(cobra.ExactArgs(1)),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logger := logger.New(config.LogVerbosity())
		client, err := plex.New(logger, config.PlexServerUrl(), config.PlexToken())
		if err != nil {
			return err
		}

		// 1. Load the cache
		if err := client.LoadCache(ctx); err != nil {
			return fmt.Errorf("error loading cache: %w", err)
		}

		// 2. Load the mediux file
		filePath := args[0]
		logger.Info(fmt.Sprintf("Loading config from '%s'", filePath))
		cfg, err := mediux.ConfigFromFile(filePath)
		if err != nil {
			return err
		}

		// 3. Loop through each library
		for _, library := range *cfg {
			logger.Infof("Processing library '%s'", library.Title)
			logger := logger.WithField("library", library.Title)

			// 4. Loop through each item
			for _, item := range library.Items {
				switch item.Type {
				case types.MediaTypeMovie:
					client.HandleMovie(ctx, logger, item)
				case types.MediaTypeShow:
					client.HandleShow(ctx, logger, item)
				default:
					logger.Errorf("Unknown item type '%s' for item %s", item.Type, item.Id)
				}
			}
		}

		return nil
	},
}

func init() {
	Cmd.AddCommand(fileCmd)
}
