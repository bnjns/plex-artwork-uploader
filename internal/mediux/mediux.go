package mediux

import (
	"fmt"
	"github.com/bnjns/plex-artwork-uploader/internal/types"
	"github.com/bnjns/plex-artwork-uploader/internal/utils"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
)

type configFile = map[string]map[int]configFileItem

type configFileItem struct {
	PosterUrl     string                    `yaml:"url_poster"`
	BackgroundUrl string                    `yaml:"url_background"`
	Seasons       *map[int]configFileSeason `yaml:"seasons,omitempty"`
}

type configFileSeason struct {
	PosterUrl     string                    `yaml:"url_poster"`
	BackgroundUrl string                    `yaml:"url_background"`
	Episodes      map[int]configFileEpisode `yaml:"episodes"`
}

type configFileEpisode struct {
	PosterUrl     string `yaml:"url_poster"`
	BackgroundUrl string `yaml:"url_background"`
}

func ConfigFromFile(filePath string) (*[]types.Library, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading mediux file: %w", err)
	}

	var file configFile
	if err = yaml.Unmarshal(fileBytes, &file); err != nil {
		return nil, fmt.Errorf("error parsing mediux file: %w", err)
	}

	cfg := make([]types.Library, 0)
	for title, items := range file {
		libraryItems := make([]types.LibraryItem, 0)

		utils.IterateOrdered(items, func(id int, item configFileItem) {
			libraryItem := types.LibraryItem{
				Id:    strconv.Itoa(id),
				Agent: types.AgentTMDB,
				Type:  types.MediaTypeMovie,
				Artwork: types.Artwork{
					PosterUrl:     item.PosterUrl,
					BackgroundUrl: item.BackgroundUrl,
				},
			}

			if item.Seasons != nil {
				libraryItem.Agent = types.AgentTVDB
				libraryItem.Type = types.MediaTypeShow

				seasons := make([]types.Season, 0)
				utils.IterateOrdered(*item.Seasons, func(seasonNum int, season configFileSeason) {
					episodes := make([]types.Episode, 0)
					utils.IterateOrdered(season.Episodes, func(episodeNum int, episode configFileEpisode) {
						episodes = append(episodes, types.Episode{
							Number: episodeNum,
							Artwork: types.Artwork{
								PosterUrl:     episode.PosterUrl,
								BackgroundUrl: episode.BackgroundUrl,
							},
						})
					})

					seasons = append(seasons, types.Season{
						Number: seasonNum,
						Artwork: types.Artwork{
							PosterUrl:     season.PosterUrl,
							BackgroundUrl: season.BackgroundUrl,
						},
						Episodes: episodes,
					})
				})
				libraryItem.Seasons = seasons
			}

			libraryItems = append(libraryItems, libraryItem)
		})

		cfg = append(cfg, types.Library{
			Title: title,
			Items: libraryItems,
		})
	}

	return &cfg, nil
}
