package crawler

import "github.com/dkotik/pageseo"

func NewTargetFilter(URLs ...string) (Filter, error) {
	return FilterFunc(func(r pageseo.Resource) bool {
		panic("impl")
		// return true
	}), nil
}
