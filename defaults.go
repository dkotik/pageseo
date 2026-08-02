package pageseo

import "slices"

const (
	DefaultMinimumTitleLength       = 4
	DefaultMaximumTitleLength       = 55
	DefaultMinimumHeadingLength     = DefaultMinimumTitleLength
	DefaultMaximumHeadingLength     = 70
	DefaultMinimumDescriptionLength = 4
	DefaultMaximumDescriptionLength = 150
)

func DefaultNodeTests() []NodeTester {
	return []NodeTester{
		NewHeadingNodeTester(StringConstraints{}),
	}
}

func DefaultNodeTestsWith(extensions ...NodeTester) []NodeTester {
	return slices.Concat(DefaultNodeTests(), extensions)
}
