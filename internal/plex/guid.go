package plex

import (
	"fmt"
	"strings"

	"github.com/bnjns/plex-artwork-uploader/internal/types"
)

func canonicalGuid(guid string) string {
	normalized := strings.TrimSpace(strings.ToLower(guid))
	return strings.TrimPrefix(normalized, "com.plexapp.agents.")
}

func inferMediaTypeFromGuid(guid string) (types.MediaType, error) {
	canonical := canonicalGuid(guid)
	switch {
	case strings.HasPrefix(canonical, "tvdb://"):
		return types.MediaTypeShow, nil
	case strings.HasPrefix(canonical, "tmdb://"):
		return types.MediaTypeMovie, nil
	default:
		return "", fmt.Errorf("unsupported guid type: %s", guid)
	}
}
