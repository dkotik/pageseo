package pageseo

import "slices"

const (
	DefaultMinimumTitleLength       = 4
	DefaultMaximumTitleLength       = 55
	DefaultMinimumHeadingLength     = DefaultMinimumTitleLength
	DefaultMaximumHeadingLength     = 70
	DefaultMinimumDescriptionLength = 4
	DefaultMaximumDescriptionLength = 150
	DefaultMaximumKeywordsLength    = DefaultMaximumDescriptionLength
)

func DefaultNodeTests() []NodeTester {
	return []NodeTester{
		NewHeadNodeTester(HeadNodeConstraints{}),
		NewHeadingNodeTester(StringConstraints{}),
	}
}

func DefaultNodeTestsWith(extensions ...NodeTester) []NodeTester {
	return slices.Concat(DefaultNodeTests(), extensions)
}
