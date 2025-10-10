package plex

import (
	"context"
	"fmt"
	"github.com/LukeHagar/plexgo/models/operations"
	"github.com/bnjns/plex-artwork-uploader/internal/types"
	"github.com/bnjns/plex-artwork-uploader/internal/utils"
	"github.com/sirupsen/logrus"
	"regexp"
	"strconv"
)

var guidRegex = regexp.MustCompile(`(?i)^(?:com\.plexapp\.agents\.)?(?P<agent>[a-z]+)://(?P<id>[a-z0-9]+)`)

func (c *Client) LoadCache(ctx context.Context) error {
	c.logger.Info("Building library and media cache")
	allLibrariesResponse, err := c.api.Library.GetAllLibraries(ctx)
	if err != nil {
		return fmt.Errorf("error loading libraries: %w", err)
	}

	mediaContainer := allLibrariesResponse.GetObject().GetMediaContainer()
	for _, library := range mediaContainer.GetDirectory() {
		libraryKey, err := strconv.Atoi(library.GetKey())
		if err != nil {
			c.logger.WithError(err).Warnf("Library %s will not be cached.", library.GetKey())
			continue
		}
		logger := c.logger.WithField("libraryKey", libraryKey)

		logger.Debugf("Caching movies for library '%s'", library.Title)
		if err := c.cacheMovies(ctx, logger, libraryKey); err != nil {
			return fmt.Errorf("error caching movies for library '%s': %w", library.Title, err)
		}

		logger.Debugf("Caching shows for library '%s'", library.Title)
		if err := c.cacheShows(ctx, logger, libraryKey); err != nil {
			return fmt.Errorf("error caching shows for library '%s': %w", library.Title, err)
		}
	}

	return nil
}

func (c *Client) cacheMovies(ctx context.Context, logger *logrus.Entry, libraryKey int) error {
	allMovies, err := c.getAllItemsInLibrary(ctx, libraryKey, operations.GetAllMediaLibraryQueryParamTypeMovie)
	if err != nil {
		return err
	}

	for _, movie := range *allMovies {
		item := &types.CacheLibraryItem{
			RatingKey: convertRatingKey(movie.GetRatingKey()),
			Title:     movie.GetTitle(),
			Type:      types.MediaTypeMovie,
		}

		matches := guidRegex.FindStringSubmatch(movie.GetGUID())

		c.cache[utils.MakeGuid(utils.NormaliseAgent(matches[1]), matches[2])] = item

		for _, guid := range movie.GetGuids() {
			c.cache[*guid.GetID()] = item
		}
	}

	logger.Debugf("Cached %d movies", len(*allMovies))
	return nil
}

func (c *Client) cacheShows(ctx context.Context, logger *logrus.Entry, libraryKey int) error {
	logger.Tracef("Fetching all shows for library %d", libraryKey)
	allShows, err := c.getAllItemsInLibrary(ctx, libraryKey, operations.GetAllMediaLibraryQueryParamTypeTvShow)
	if err != nil {
		return fmt.Errorf("error fetching shows: %w", err)
	}

	logger.Tracef("Fetching all seasons for library %d", libraryKey)
	allSeasons, err := c.getAllItemsInLibrary(ctx, libraryKey, operations.GetAllMediaLibraryQueryParamTypeSeason)
	if err != nil {
		return fmt.Errorf("error fetching seasons: %w", err)
	}

	logger.Tracef("Fetching all episodes for library %d", libraryKey)
	allEpisodes, err := c.getAllItemsInLibrary(ctx, libraryKey, operations.GetAllMediaLibraryQueryParamTypeEpisode)
	if err != nil {
		return fmt.Errorf("error fetching seasons: %w", err)
	}

	for _, show := range *allShows {
		seasons := make(map[int]types.CacheLibraryItemSeason)

		for _, season := range *allSeasons {
			if season.GetParentRatingKey() != nil && *season.GetParentRatingKey() == show.GetRatingKey() {
				episodes := make(map[int]types.CacheLibraryItemEpisode)

				for _, episode := range *allEpisodes {
					if episode.GetParentRatingKey() != nil && *episode.GetParentRatingKey() == season.GetRatingKey() {
						episodes[episode.GetIndex()] = types.CacheLibraryItemEpisode{
							RatingKey: convertRatingKey(episode.GetRatingKey()),
						}
					}
				}

				seasons[season.GetIndex()] = types.CacheLibraryItemSeason{
					RatingKey: convertRatingKey(season.GetRatingKey()),
					Episodes:  episodes,
				}
			}
		}

		cacheItem := types.CacheLibraryItem{
			RatingKey: convertRatingKey(show.GetRatingKey()),
			Title:     show.GetTitle(),
			Type:      types.MediaTypeShow,
			Seasons:   seasons,
		}

		for _, guid := range show.GetGuids() {
			c.cache[*guid.GetID()] = &cacheItem
		}
	}

	logger.Debugf("%d shows, %d seasons and %d episodes cached", len(*allShows), len(*allSeasons), len(*allEpisodes))
	return nil
}

func (c *Client) getAllItemsInLibrary(ctx context.Context, libraryKey int, mediaType operations.GetAllMediaLibraryQueryParamType) (*[]operations.GetAllMediaLibraryMetadata, error) {
	items := make([]operations.GetAllMediaLibraryMetadata, 0)
	page := 0
	containerSize := 100
	for {
		c.logger.Tracef("Fetching page %d of type %d in library %d", page+1, mediaType, libraryKey)
		containerStart := page * containerSize
		response, err := c.api.Library.GetAllMediaLibrary(ctx, operations.GetAllMediaLibraryRequest{
			SectionKey:          libraryKey,
			Type:                mediaType,
			IncludeGuids:        operations.QueryParamIncludeGuidsEnable.ToPointer(),
			XPlexContainerStart: &containerStart,
			XPlexContainerSize:  &containerSize,
		})
		if err != nil {
			return nil, fmt.Errorf("error loading page %d for library %d and media type %d: %w", containerStart+1, libraryKey, mediaType, err)
		}

		for _, item := range response.GetObject().GetMediaContainer().GetMetadata() {
			items = append(items, item)
		}

		if response.GetObject().GetMediaContainer().GetSize() < containerSize {
			break
		}

		page++
	}

	c.logger.Tracef("Found a total of %d items across %d pages", len(items), page+1)
	return &items, nil
}
