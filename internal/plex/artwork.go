package plex

import (
	"context"
	"strconv"

	"github.com/LukeHagar/plexgo/models/operations"
)

func (c *Client) UploadPoster(ctx context.Context, ratingKey int64, url string) error {
	_, err := c.api.Library.SetItemArtwork(ctx, operations.SetItemArtworkRequest{
		Ids:     strconv.FormatInt(ratingKey, 10),
		Element: operations.Element("posters"),
		URL:     &url,
	})
	return err
}

func (c *Client) UploadBackground(ctx context.Context, ratingKey int64, url string) error {
	_, err := c.api.Library.SetItemArtwork(ctx, operations.SetItemArtworkRequest{
		Ids:     strconv.FormatInt(ratingKey, 10),
		Element: operations.Element("arts"),
		URL:     &url,
	})
	return err
}
