package manual

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bnjns/plex-artwork-uploader/internal/config"
	"github.com/bnjns/plex-artwork-uploader/internal/logger"
	"github.com/bnjns/plex-artwork-uploader/internal/plex"
	"github.com/bnjns/plex-artwork-uploader/internal/types"
	"github.com/bnjns/plex-artwork-uploader/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	flagRatingKey     string
	flagPosterUrl     string
	flagBackgroundUrl string

	flagMovieId string

	flagShowId      string
	flagShowSeason  string
	flagShowEpisode string
)

var Cmd = &cobra.Command{
	Use:   "manual",
	Short: "Manually provide the details of the artwork and movie/show",
	Long: `This command allows you to manually upload artwork to a movie or show.

There are 2 ways to specify the movie/show:

1. Manually provide the "rating key" with the --rating-key flag

plex-artwork-uploader manual --rating-key 12345 [--poster-url <url>] [--background-url <url>]

This bypasses any library checks so this will simply error if the movie/show doesn't exist.

2. Provide the TMDB/TVDB ID (and optionally season and episode for shows):

plex-artwork-uploader manual --movie-id 12345 [--poster-url <url>] [--background-url <url>]
plex-artwork-uploader manual --show-id 12345 [--poster-url <url>] [--background-url <url>]
plex-artwork-uploader manual --show-id 12345 --season 1 [--poster-url <url>] [--background-url <url>]
plex-artwork-uploader manual --show-id 12345 --season 1 --episode 1 [--poster-url <url>] [--background-url <url>]

For both movies and shows, this will attempt to convert the TMDB/TVDB ID into the rating key before uploading and will
error if it's not found.

For shows, you can:

- Upload the show poster/background by just providing --show-id
- Upload a season poster/background by providing --show-id and --season
- Upload an episode poster/background by providing --show-id, --season and --episode
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		logger := logger.New(config.LogVerbosity())

		client, err := plex.New(logger, config.PlexServerUrl(), config.PlexToken())
		if err != nil {
			return err
		}

		artwork := types.Artwork{
			PosterUrl:     flagPosterUrl,
			BackgroundUrl: flagBackgroundUrl,
		}

		if flagRatingKey != "" {
			ratingKey, _ := strconv.ParseInt(flagRatingKey, 10, 64)
			err := client.UploadArtwork(ctx, logger.WithContext(ctx), ratingKey, fmt.Sprintf("item %d", ratingKey), artwork)
			if err != nil {
				logger.Error(err.Error())
			}
		} else {
			if err := client.LoadCache(ctx); err != nil {
				return fmt.Errorf("error loading cache: %w", err)
			}

			if flagMovieId != "" {
				client.HandleMovie(ctx, logger.WithContext(ctx), types.LibraryItem{
					Id:      flagMovieId,
					Agent:   types.AgentTMDB,
					Type:    types.MediaTypeMovie,
					Artwork: artwork,
				})
			} else if flagShowId != "" {
				err := processShow(ctx, client, logger.WithContext(ctx), flagShowId, flagShowSeason, flagShowEpisode, artwork)
				if err != nil {
					logger.Error(err.Error())
				}
			}
		}

		return nil
	},
}

func processShow(ctx context.Context, client *plex.Client, logger *logrus.Entry, showId string, seasonNumStr string, episodeNumStr string, artwork types.Artwork) error {
	showGuid := utils.MakeGuid(string(types.AgentTVDB), showId)
	logger = logger.WithField("showGuid", showGuid)

	if seasonNumStr != "" && episodeNumStr != "" {
		seasonNum, _ := strconv.Atoi(seasonNumStr)
		episodeNum, _ := strconv.Atoi(episodeNumStr)
		episode, err := client.FindEpisode(ctx, showGuid, seasonNum, episodeNum)
		if err != nil {
			return err
		}

		return client.UploadArtwork(ctx, logger, episode.RatingKey, fmt.Sprintf("S%02dE%02d", seasonNum, episodeNum), artwork)
	} else if seasonNumStr != "" {
		seasonNum, _ := strconv.Atoi(seasonNumStr)
		season, err := client.FindSeason(ctx, showGuid, seasonNum)
		if err != nil {
			return err
		}

		return client.UploadArtwork(ctx, logger, season.RatingKey, fmt.Sprintf("S%02d", seasonNum), artwork)
	} else {
		item, err := client.FindItemForGuid(ctx, showGuid)
		if err != nil {
			return fmt.Errorf("error finding show with guid %s: %w", showGuid, err)
		}

		logger = logger.WithField("show", item.Title)
		return client.UploadArtwork(ctx, logger, item.RatingKey, fmt.Sprintf("show '%s'", item.Title), artwork)
	}
}

func init() {
	Cmd.Flags().StringVar(&flagPosterUrl, "poster-url", "", "The url of the poster image to upload.")
	Cmd.Flags().StringVar(&flagBackgroundUrl, "background-url", "", "The url of the background image to upload.")

	Cmd.Flags().StringVar(&flagRatingKey, "rating-key", "", "The rating key of the movie or show to upload the artwork to.")

	Cmd.Flags().StringVar(&flagMovieId, "movie-id", "", "The TMDB ID of the movie to upload the artwork to.")

	Cmd.Flags().StringVar(&flagShowId, "show-id", "", "The TVDB ID of the show to upload the artwork to.")
	Cmd.Flags().StringVar(&flagShowSeason, "season", "", "The season number to upload the artwork to. Requires --show-id to be set.")
	Cmd.Flags().StringVar(&flagShowEpisode, "episode", "", "The episode number to upload the artwork to. Requires --season to be set.")

	Cmd.MarkFlagsMutuallyExclusive("rating-key", "movie-id", "show-id")
	Cmd.MarkFlagsOneRequired("rating-key", "movie-id", "show-id")
}
