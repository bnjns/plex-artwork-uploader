package plex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/LukeHagar/plexgo/models/components"
)

type librarySection struct {
	Key   string
	Type  string
	Title string
}

type sourceMediaItem struct {
	Title           string
	RatingKey       string
	GUID            string
	ParentRatingKey string
	Index           *int
	Guids           []string
}

// httpClient performs direct HTTP calls to the Plex server (avoids SDK JSON models where responses drift).
type httpClient struct {
	serverURL string
	token     string
}

func (s *httpClient) GetLibrarySections(ctx context.Context) ([]librarySection, error) {
	u, err := url.Parse(s.serverURL + "/library/sections")
	if err != nil {
		return nil, fmt.Errorf("error parsing sections URL: %w", err)
	}

	var payload struct {
		MediaContainer struct {
			Directory []struct {
				Key   string `json:"key"`
				Type  string `json:"type"`
				Title string `json:"title"`
			} `json:"Directory"`
		} `json:"MediaContainer"`
	}
	if err := s.doJSON(ctx, http.MethodGet, u.String(), &payload); err != nil {
		return nil, err
	}

	sections := make([]librarySection, 0, len(payload.MediaContainer.Directory))
	for _, section := range payload.MediaContainer.Directory {
		if section.Key == "" || section.Type == "" {
			continue
		}
		sections = append(sections, librarySection{
			Key:   section.Key,
			Type:  section.Type,
			Title: section.Title,
		})
	}
	return sections, nil
}

func (s *httpClient) ListSectionContent(ctx context.Context, sectionID string, mediaType components.MediaType, start int, size int) ([]sourceMediaItem, error) {
	u, err := url.Parse(fmt.Sprintf("%s/library/sections/%s/all", s.serverURL, sectionID))
	if err != nil {
		return nil, fmt.Errorf("error parsing content URL: %w", err)
	}
	query := u.Query()
	query.Set("X-Plex-Container-Start", strconv.Itoa(start))
	query.Set("X-Plex-Container-Size", strconv.Itoa(size))
	query.Set("type", strconv.FormatInt(int64(mediaType), 10))
	query.Set("includeGuids", "1")
	u.RawQuery = query.Encode()

	var payload struct {
		MediaContainer struct {
			Metadata []struct {
				Title           string `json:"title"`
				RatingKey       string `json:"ratingKey"`
				GUID            string `json:"guid"`
				ParentRatingKey string `json:"parentRatingKey"`
				Index           *int   `json:"index"`
				Guids           []struct {
					ID string `json:"id"`
				} `json:"Guid"`
			} `json:"Metadata"`
		} `json:"MediaContainer"`
	}
	if err := s.doJSON(ctx, http.MethodGet, u.String(), &payload); err != nil {
		return nil, err
	}

	items := make([]sourceMediaItem, 0, len(payload.MediaContainer.Metadata))
	for _, raw := range payload.MediaContainer.Metadata {
		item := sourceMediaItem{
			Title:           raw.Title,
			RatingKey:       raw.RatingKey,
			GUID:            raw.GUID,
			ParentRatingKey: raw.ParentRatingKey,
			Index:           raw.Index,
		}
		for _, g := range raw.Guids {
			if g.ID == "" {
				continue
			}
			item.Guids = append(item.Guids, g.ID)
		}
		items = append(items, item)
	}

	return items, nil
}

func (s *httpClient) doJSON(ctx context.Context, method string, rawURL string, out any) error {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if s.token != "" {
		req.Header.Set("X-Plex-Token", s.token)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("error calling plex endpoint: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("plex endpoint returned status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return fmt.Errorf("error decoding plex JSON response: %w", err)
	}
	return nil
}
