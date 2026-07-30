package main

import (
	"net/url"
	"slices"
	"strings"
	"sync"

	"github.com/dkotik/pageseo"
)

type resourceQueue struct {
	mu      sync.Mutex
	targets []pageseo.Resource
	domains []string
	// limit   int
}

func newResourceQueue(origins ...string) *resourceQueue {
	domains := []string{}
	for _, origin := range origins {
		parsed, err := url.Parse(origin)
		if err == nil {
			host := strings.TrimSpace(parsed.Hostname())
			if host != "" && slices.Index(domains, host) == -1 {
				domains = append(domains, host)
			}
		}
	}
	// if limit == 0 {
	// 	limit = 1
	// }
	return &resourceQueue{
		mu:      sync.Mutex{},
		targets: make([]pageseo.Resource, 0, 64),
		domains: domains,
		// limit:   int(limit),
	}
}

func (rq *resourceQueue) Push(r pageseo.Resource) {
	if r.Error != nil || r.ContentType != "text/html" {
		// fmt.Println("SKIP ", r.URL, r.Error)
		return
	}
	parsed, err := url.Parse(r.URL)
	if err != nil {
		return
	}
	host := strings.TrimSpace(parsed.Hostname())
	if host == "" || slices.Index(rq.domains, host) == -1 {
		return
	}

	rq.mu.Lock()
	rq.targets = append(rq.targets, r)
	rq.mu.Unlock()
}

func (rq *resourceQueue) Pull() (result []pageseo.Resource) {
	rq.mu.Lock()
	result = slices.Clone(rq.targets)
	rq.targets = rq.targets[:0]
	rq.mu.Unlock()
	return result
}
