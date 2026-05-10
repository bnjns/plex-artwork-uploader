package plex

import "testing"

func TestBuildShowHierarchyFromSectionData(t *testing.T) {
	show := sourceMediaItem{
		Title:     "The Rookie",
		RatingKey: "22959",
	}
	season := sourceMediaItem{
		RatingKey:       "22960",
		ParentRatingKey: "22959",
		Index:           intPtr(1),
	}
	episode := sourceMediaItem{
		RatingKey:       "22961",
		ParentRatingKey: "22960",
		Index:           intPtr(3),
	}

	item := buildShowHierarchyFromSectionData(show, []sourceMediaItem{season}, []sourceMediaItem{episode})
	if item.Title != "The Rookie" {
		t.Fatalf("unexpected show title: %q", item.Title)
	}
	s1, ok := item.Seasons[1]
	if !ok {
		t.Fatalf("expected season 1 in hierarchy")
	}
	e3, ok := s1.Episodes[3]
	if !ok {
		t.Fatalf("expected episode 3 in season 1 hierarchy")
	}
	if e3.RatingKey != 22961 {
		t.Fatalf("unexpected episode rating key: %d", e3.RatingKey)
	}
}

func intPtr(v int) *int {
	return &v
}
