package plex

import (
	"errors"
	"github.com/LukeHagar/plexgo"
	"github.com/bnjns/plex-artwork-uploader/internal/types"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound = errors.New("not found")
)

type Client struct {
	api    *plexgo.PlexAPI
	logger *logrus.Logger

	cache map[string]*types.CacheLibraryItem
}

func New(logger *logrus.Logger, serverUrl string, token string) (*Client, error) {
	client := &Client{
		api:    plexgo.New(plexgo.WithServerURL(serverUrl), plexgo.WithSecurity(token)),
		logger: logger,

		cache: make(map[string]*types.CacheLibraryItem),
	}

	return client, nil
}
