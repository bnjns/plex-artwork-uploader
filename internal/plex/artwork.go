package plex

import (
	"context"
)

func (c *Client) UploadPoster(ctx context.Context, ratingKey int64, url string) error {
	_, err := c.api.Library.PostMediaPoster(ctx, ratingKey, &url, nil)
	return err
}

func (c *Client) UploadBackground(ctx context.Context, ratingKey int64, url string) error {
	_, err := c.api.Library.PostMediaArts(ctx, ratingKey, &url, nil)
	return err
}
