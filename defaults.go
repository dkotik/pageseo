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
	DefaultMaximumURLLength          = 2048 // older browser constraint
	DefaultMinimumAnchorTextLength   = 1
	DefaultMaximumAnchorTextLength   = DefaultMaximumTitleLength * 6
)

func DefaultNodeTests() []NodeTester {
	return []NodeTester{
		NewHeadNodeTester(HeadNodeConstraints{}),
		NewHeadingNodeTester(StringConstraints{}),
		NewAnchorNodeTester(StringConstraints{}),
		NewImageNodeTester(StringConstraints{}),
		NewScriptNodeTester(),
	}
}

func DefaultNodeTestsWith(extensions ...NodeTester) []NodeTester {
	return slices.Concat(DefaultNodeTests(), extensions)
}
