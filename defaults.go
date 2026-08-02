package pageseo

const (
	DefaultMinimumTitleLength       = 4
	DefaultMaximumTitleLength       = 55
	DefaultMinimumHeadingLength     = DefaultMinimumTitleLength
	DefaultMaximumHeadingLength     = DefaultMaximumTitleLength
	DefaultMinimumDescriptionLength = 4
	DefaultMaximumDescriptionLength = 150
)

func GetDefaultNodeTests() []NodeTester {
	return []NodeTester{}
}
