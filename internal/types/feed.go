package types

import (
	"fmt"

	ierr "github.com/omkar273/nashikdarshan/internal/errors"
	"github.com/samber/lo"
)

// FeedSectionType represents the type of feed section
type FeedSectionType string

const (
	SectionTypeLatest   FeedSectionType = "latest"
	SectionTypeTrending FeedSectionType = "trending"
	SectionTypePopular  FeedSectionType = "popular"
	SectionTypeNearby   FeedSectionType = "nearby"
)

// FeedSectionTypes contains all valid feed section types
var FeedSectionTypes = []string{
	string(SectionTypeLatest),
	string(SectionTypeTrending),
	string(SectionTypePopular),
	string(SectionTypeNearby),
}

func (f FeedSectionType) Validate() error {
	if !lo.Contains(FeedSectionTypes, string(f)) {
		return ierr.NewError("invalid section type").
			WithHint(fmt.Sprintf("invalid section type: %s. Valid types are: %v", f, FeedSectionTypes)).
			Mark(ierr.ErrValidation)
	}
	return nil
}
