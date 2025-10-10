package plex

import "strconv"

func convertRatingKey(ratingKey string) int64 {
	ratingKeyInt, _ := strconv.ParseInt(ratingKey, 10, 64)
	return ratingKeyInt
}
