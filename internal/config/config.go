package config

import (
	"github.com/bnjns/plex-artwork-uploader/internal/flags"
	"os"
)

const (
	envPlexServerUrl = "PLEX_SERVER_URL"
	envPlexToken     = "PLEX_TOKEN"
)

func LogVerbosity() int {
	return flags.LogVerbosity
}

func PlexServerUrl() string {
	if flags.PlexServerUrl != "" {
		return flags.PlexServerUrl
	}

	return os.Getenv(envPlexServerUrl)
}

func PlexToken() string {
	if flags.PlexToken != "" {
		return flags.PlexToken
	}

	return os.Getenv(envPlexToken)
}
