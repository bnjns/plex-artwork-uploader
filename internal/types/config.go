package types

import (
	"github.com/bnjns/plex-artwork-uploader/internal/utils"
)

type Library struct {
	Title string
	Items []LibraryItem
}

type LibraryItem struct {
	Id    string
	Agent Agent
	Type  MediaType
	Artwork

	Seasons []Season
}

type Season struct {
	Number int
	Artwork
	Episodes []Episode
}

type Episode struct {
	Number int
	Artwork
}

type Artwork struct {
	PosterUrl     string
	BackgroundUrl string
}

func (i LibraryItem) MakeGuid() string {
	return utils.MakeGuid(string(i.Agent), i.Id)
}
