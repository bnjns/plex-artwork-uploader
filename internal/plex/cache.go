package plex

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/LukeHagar/plexgo/models/components"
	"github.com/bnjns/plex-artwork-uploader/internal/types"
	"github.com/bnjns/plex-artwork-uploader/internal/utils"
)

var guidRegex = regexp.MustCompile(`(?i)^(?:com\.plexapp\.agents\.)?(?P<agent>[a-z]+)://(?P<id>[a-z0-9]+)`)

func (c *Client) LoadCache(ctx context.Context) error {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	if c.cacheLoaded {
		c.logger.Debugf("Cache already loaded (%d keys)", len(c.cache))
		return nil
	}

	startedAt := time.Now()
	c.logger.Info("Building library cache")
	newCache := make(map[string]*types.CacheLibraryItem)
	sections, err := c.httpClient.GetLibrarySections(ctx)
	if err != nil {
		return fmt.Errorf("error loading library sections: %w", err)
	}

	movieSectionCount := 0
	showSectionCount := 0
	for _, section := range sections {
		switch section.Type {
		case "movie":
			movieSectionCount++
			if err := c.cacheMovies(ctx, section.Key, newCache); err != nil {
				return fmt.Errorf("error caching movies for library '%s': %w", section.Title, err)
			}
		case "show":
			showSectionCount++
			if err := c.cacheShows(ctx, section.Key, newCache); err != nil {
				return fmt.Errorf("error caching shows for library '%s': %w", section.Title, err)
			}
		}
	}

	c.cache = newCache
	c.cacheLoaded = true

	c.logger.Infof("Library cache built in %s (%d keys, %d sections, %d movie sections, %d show sections)", time.Since(startedAt).String(), len(newCache), len(sections), movieSectionCount, showSectionCount)

	return nil
}

func (c *Client) cacheMovies(ctx context.Context, sectionID string, cache map[string]*types.CacheLibraryItem) error {
	allMovies, err := c.getAllItemsInLibrary(ctx, sectionID, components.MediaTypeMovie)
	if err != nil {
		return err
	}

	for _, movie := range allMovies {
		item := &types.CacheLibraryItem{
			RatingKey: convertRatingKey(movie.RatingKey),
			Title:     movie.Title,
			Type:      types.MediaTypeMovie,
		}
		addItemToCache(cache, movie, item)
	}

	c.logger.Debugf("Cached %d movies from section %s", len(allMovies), sectionID)
	return nil
}

func (c *Client) cacheShows(ctx context.Context, sectionID string, cache map[string]*types.CacheLibraryItem) error {
	allShows, err := c.getAllItemsInLibrary(ctx, sectionID, components.MediaTypeTvShow)
	if err != nil {
		return fmt.Errorf("error fetching shows: %w", err)
	}

	allSeasons, err := c.getAllItemsInLibrary(ctx, sectionID, components.MediaTypeSeason)
	if err != nil {
		return fmt.Errorf("error fetching seasons: %w", err)
	}

	allEpisodes, err := c.getAllItemsInLibrary(ctx, sectionID, components.MediaTypeEpisode)
	if err != nil {
		return fmt.Errorf("error fetching episodes: %w", err)
	}

	for _, show := range allShows {
		cacheItem := buildShowHierarchyFromSectionData(show, allSeasons, allEpisodes)
		addItemToCache(cache, show, cacheItem)
	}

	c.logger.Debugf("Cached %d shows, %d seasons and %d episodes from section %s", len(allShows), len(allSeasons), len(allEpisodes), sectionID)
	return nil
}

func (c *Client) getAllItemsInLibrary(ctx context.Context, sectionID string, mediaType components.MediaType) ([]sourceMediaItem, error) {
	items := make([]sourceMediaItem, 0)
	page := 0
	containerSize := 200
	for {
		c.logger.Tracef("Fetching page %d of type %d in library %s", page+1, mediaType, sectionID)
		containerStart := page * containerSize
		pageItems, err := c.httpClient.ListSectionContent(ctx, sectionID, mediaType, containerStart, containerSize)
		if err != nil {
			return nil, fmt.Errorf("error loading page %d for library %s and media type %d: %w", containerStart+1, sectionID, mediaType, err)
		}
		items = append(items, pageItems...)
		if len(pageItems) < containerSize {
			break
		}
		page++
	}

	return items, nil
}

func addItemToCache(cache map[string]*types.CacheLibraryItem, item sourceMediaItem, cacheItem *types.CacheLibraryItem) {
	if item.GUID != "" {
		cache[canonicalGuid(item.GUID)] = cacheItem
		if matches := guidRegex.FindStringSubmatch(strings.ToLower(item.GUID)); len(matches) == 3 {
			cache[utils.MakeGuid(utils.NormaliseAgent(matches[1]), matches[2])] = cacheItem
		}
	}

	for _, guid := range item.Guids {
		if guid != "" {
			cache[canonicalGuid(guid)] = cacheItem
		}
	}
}
