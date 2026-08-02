package pageseo

import "slices"

const (
	DefaultMinimumTitleLength        = 4
	DefaultMaximumTitleLength        = 55
	DefaultMinimumHeadingLength      = DefaultMinimumTitleLength
	DefaultMaximumHeadingLength      = 70
	DefaultMinimumDescriptionLength  = 4
	DefaultMaximumDescriptionLength  = 150
	DefaultMaximumKeywordsLength     = DefaultMaximumDescriptionLength
	DefaultMinimumImageAltTextLength = 0
	DefaultMaximumImageAltTextLength = DefaultMaximumDescriptionLength

	// DefaultMinimumLinkTextLength sets the minimum length of the anchor text.
	// A pagination link is often a single character.
	DefaultMinimuLinkLength         = 1
	DefaultMaximumLinkLength        = 120
	DefaultMaximumImageSourceLength = 2048 // older browser constraint
	DefaultMinimumLinkTextLength    = 1
	DefaultMaximumLinkTextLength    = DefaultMaximumTitleLength * 6
)

func DefaultNodeTests() []NodeTester {
	return []NodeTester{
		NewHeadNodeTester(HeadNodeConstraints{}),
		NewHeadingNodeTester(StringConstraints{}),
		NewImageNodeTester(StringConstraints{}),
		NewScriptNodeTester(),
	}
}

func DefaultNodeTestsWith(extensions ...NodeTester) []NodeTester {
	return slices.Concat(DefaultNodeTests(), extensions)
}
