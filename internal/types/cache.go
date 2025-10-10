package types

type CacheLibraryItem struct {
	RatingKey int64
	Title     string
	Type      MediaType
	Seasons   map[int]CacheLibraryItemSeason
}

type CacheLibraryItemSeason struct {
	RatingKey int64
	Episodes  map[int]CacheLibraryItemEpisode
}

type CacheLibraryItemEpisode struct {
	RatingKey int64
}
