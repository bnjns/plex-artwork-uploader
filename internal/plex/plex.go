package plex

import (
	"errors"
	"strings"
	"sync"

	"github.com/LukeHagar/plexgo"
	"github.com/LukeHagar/plexgo/models/components"
	"github.com/bnjns/plex-artwork-uploader/internal/types"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrCacheNotLoaded = errors.New("plex library cache not loaded; call LoadCache first")
)

type Client struct {
	api        *plexgo.PlexAPI
	httpClient *httpClient
	logger     *logrus.Logger

	cache       map[string]*types.CacheLibraryItem
	cacheLoaded bool
	cacheMu     sync.Mutex
}

func New(logger *logrus.Logger, serverUrl string, token string) (*Client, error) {
	client := &Client{
		api:    plexgo.New(plexgo.WithServerURL(serverUrl), plexgo.WithAccepts(components.AcceptsApplicationJSON), plexgo.WithSecurity(token)),
		logger: logger,
		httpClient: &httpClient{
			serverURL: strings.TrimRight(serverUrl, "/"),
			token:     token,
		},

		cache:       make(map[string]*types.CacheLibraryItem),
		cacheLoaded: false,
	}

	return client, nil
}
