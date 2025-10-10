package plex

import (
	"context"
	"errors"
	"fmt"
	"github.com/bnjns/plex-artwork-uploader/internal/types"
	"github.com/sirupsen/logrus"
)

func (c *Client) FindItemForGuid(guid string) (*types.CacheLibraryItem, error) {
	item, exists := c.cache[guid]
	if !exists {
		return nil, ErrNotFound
	}

	return item, nil
}

func (c *Client) FindSeason(showGuid string, seasonNum int) (*types.CacheLibraryItemSeason, error) {
	item, err := c.FindItemForGuid(showGuid)
	if err != nil {
		return nil, fmt.Errorf("error fetching show %s: %w", showGuid, err)
	}

	season, exists := item.Seasons[seasonNum]
	if !exists {
		return nil, fmt.Errorf("error fetching season %d: %w", seasonNum, ErrNotFound)
	}

	return &season, nil
}

func (c *Client) FindEpisode(showGuid string, seasonNum int, episodeNum int) (*types.CacheLibraryItemEpisode, error) {
	season, err := c.FindSeason(showGuid, seasonNum)
	if err != nil {
		return nil, err
	}

	episode, exists := season.Episodes[episodeNum]
	if !exists {
		return nil, fmt.Errorf("error fetching episode %d for season %d: %w", episodeNum, seasonNum, ErrNotFound)
	}

	return &episode, nil
}

func (c *Client) UploadArtwork(ctx context.Context, logger *logrus.Entry, ratingKey int64, title string, artwork types.Artwork) error {
	var errs []error
	logger = logger.WithField("ratingKey", ratingKey)

	if artwork.PosterUrl != "" {
		logger.Debugf("Uploading poster from %s", artwork.PosterUrl)
		if err := c.UploadPoster(ctx, ratingKey, artwork.PosterUrl); err != nil {
			logger.Errorf("Error uploading poster for %s: %s", title, err.Error())
			errs = append(errs, fmt.Errorf("error uploading poster for %s: %w", title, err))
		} else {
			logger.Infof("Poster uploaded for %s", title)
		}
	}

	if artwork.BackgroundUrl != "" {
		logger.Debugf("Uploading background from %s", artwork.PosterUrl)
		if err := c.UploadBackground(ctx, ratingKey, artwork.BackgroundUrl); err != nil {
			logger.Errorf("Error uploading background for %s: %s", title, err.Error())
			errs = append(errs, fmt.Errorf("error uploading poster for %s: %w", title, err))
		} else {
			logger.Infof("Background uploaded for %s", title)
		}
	}

	return errors.Join(errs...)
}

func (c *Client) handleItem(ctx context.Context, logger *logrus.Entry, configItem types.LibraryItem) (*logrus.Entry, error) {
	guid := configItem.MakeGuid()
	logger = logger.WithField("guid", guid)

	libraryItem, err := c.FindItemForGuid(guid)
	if err != nil {
		logger.Errorf("Error finding %s with guid %s: %s", configItem.Type, guid, err.Error())
		return nil, err
	}

	logger = logger.WithField(string(configItem.Type), libraryItem.Title)

	logger.Debugf("Rating key for %s %s is %d", configItem.Type, guid, libraryItem.RatingKey)
	logger = logger.WithField("ratingKey", libraryItem.RatingKey)
	err = c.UploadArtwork(ctx, logger, libraryItem.RatingKey, fmt.Sprintf("%s '%s'", configItem.Type, libraryItem.Title), configItem.Artwork)
	return logger, err
}

func (c *Client) HandleMovie(ctx context.Context, logger *logrus.Entry, configItem types.LibraryItem) {
	_, _ = c.handleItem(ctx, logger, configItem)
}

func (c *Client) HandleShow(ctx context.Context, logger *logrus.Entry, configItem types.LibraryItem) {
	logger, _ = c.handleItem(ctx, logger, configItem)

	logger.Debugf("Show has %d seasons", len(configItem.Seasons))
	for _, configSeason := range configItem.Seasons {
		librarySeason, err := c.FindSeason(configItem.MakeGuid(), configSeason.Number)
		if err != nil {
			logger.Error(err.Error())
			continue
		}

		_ = c.UploadArtwork(ctx, logger, librarySeason.RatingKey, fmt.Sprintf("S%02d", configSeason.Number), configSeason.Artwork)

		logger.Debugf("Season %d has %d episodes", configSeason.Number, len(configSeason.Episodes))
		for _, configEpisode := range configSeason.Episodes {
			libraryEpisode, err := c.FindEpisode(configItem.MakeGuid(), configSeason.Number, configEpisode.Number)
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			_ = c.UploadArtwork(ctx, logger, libraryEpisode.RatingKey, fmt.Sprintf("S%02dE%02d", configSeason.Number, configEpisode.Number), configEpisode.Artwork)
		}
	}
}
