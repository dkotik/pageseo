package pageseo

import (
	"context"
	"iter"
	"net/url"
	"path"
	"testing"

	"github.com/dkotik/pageseo/htmltest"
	"golang.org/x/net/html"
)

type Resource struct {
	URL         string
	ContentType string
	Content     []byte
	Error       error
}

type resourceAssociation struct {
	Node       *html.Node
	Resource   Resource
	IsExternal bool
}

func newResourceAssociation(
	node *html.Node,
	origin *url.URL,
	URL *url.URL,
) resourceAssociation {
	URL.Path = path.Clean(URL.Path)
	isExternal := IsExternalLocation(origin, URL)
	if !isExternal {
		URL = joinInternalPath(origin, URL)
	}
	return resourceAssociation{
		Node: node,
		Resource: Resource{
			URL: URL.String(),
		},
		IsExternal: isExternal,
	}
}

func (r PageValidator) enumerateResources(
	t *testing.T,
	origin *url.URL,
	node *html.Node,
) iter.Seq[resourceAssociation] {
	return func(yield func(resourceAssociation) bool) {
		var incomplete resourceAssociation
		for node := range node.Descendants() {
			if node.Type != html.ElementNode {
				continue
			}
			switch node.Data {
			case "a":
				href, ok := getAttribute(node, `href`)
				if !ok {
					t.Error(htmltest.Path(node)+":", "here is no a[href] attribute")
					continue
				}
				url, err := url.Parse(href)
				if err != nil {
					t.Errorf("%s: broken a[href] <%s>: %v", htmltest.Path(node), href, err)
					continue
				}
				incomplete = newResourceAssociation(node, origin, url)
				if !yield(incomplete) {
					return
				}
			case "img":
				src, ok := getAttribute(node, `src`)
				if !ok {
					t.Error(htmltest.Path(node)+":", "here is no img[src] attribute")
					continue
				}
				url, err := url.Parse(src)
				if err != nil {
					t.Errorf("%s: broken img[src] <%s>: %v", htmltest.Path(node), src, err)
					continue
				}
				incomplete = newResourceAssociation(node, origin, url)
				if !yield(incomplete) {
					return
				}
				// case "script":
				// 	t.Run("script tag has valid attributes", r.TestScript(node))
				// case "style":
				// 	t.Run("style tag has valid attributes", r.TestStyle(node))
			}
		}
	}
}

func (r PageValidator) loadResource(
	ctx context.Context,
	promise chan<- resourceAssociation,
	incomplete resourceAssociation,
) {
	incomplete.Resource.Content, incomplete.Resource.ContentType, incomplete.Resource.Error = r.Loader.Load(ctx, incomplete.Resource.URL)
	select {
	case <-ctx.Done():
	case promise <- incomplete:
	}
	close(promise)
}

func (r PageValidator) loadResources(
	t *testing.T,
	origin *url.URL,
	node *html.Node,
) <-chan resourceAssociation {
	results := make(chan resourceAssociation)
	var promises []chan resourceAssociation
	var promise chan resourceAssociation

	ctx := t.Context()
	for incomplete := range r.enumerateResources(t, origin, node) {
		promise = make(chan resourceAssociation)
		promises = append(promises, promise)
		go r.loadResource(ctx, promise, incomplete)
	}

	go func() {
		defer close(results)
		for _, promise := range promises {
			select {
			case <-ctx.Done():
				return
			case rc := <-promise:
				select {
				case <-ctx.Done():
					return
				case results <- rc:
				}
			}
		}
	}()
	return results
}
