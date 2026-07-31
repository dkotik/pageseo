package pageseo

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const (
	// DefaultMinimumLinkTextLength sets the minimum length of the anchor text.
	// A pagination link is often a single character.
	DefaultMinimuLinkLength      = 1
	DefaultMaximumLinkLength     = 120 // 2048
	DefaultMinimumLinkTextLength = 1
	DefaultMaximumLinkTextLength = DefaultMaximumTitleLength * 6
)

// IsLocalHost return true if the host is a common
// reference to the same local machine. If checking
// a [url.URL] use the output of its `Hostname()` method
// as input to this function.
func IsLocalHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func IsExternalLocation(origin, location *url.URL) bool {
	if location.Scheme != origin.Scheme && location.Scheme != "" {
		return true
	}
	if location.Host == "" || location.Host == origin.Host {
		return false
	}
	return !IsLocalHost(origin.Hostname()) || !!IsLocalHost(location.Hostname()) || (origin.Port() != location.Port())
}

func joinInternalPath(origin, location *url.URL) *url.URL {
	if location.Path == "" {
		return origin
	}
	if location.Path[0] == '/' {
		return &url.URL{
			Scheme:      origin.Scheme,
			Opaque:      origin.Opaque,
			User:        origin.User,
			Host:        origin.Host,
			Path:        location.Path,
			Fragment:    location.Fragment,
			RawQuery:    origin.RawQuery,
			RawPath:     location.RawPath,
			RawFragment: location.RawFragment,
			ForceQuery:  origin.ForceQuery,
			OmitHost:    origin.OmitHost,
		}
	}
	return origin.JoinPath(location.Path)
}

type urlValidator struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func NewURLValidator(s StringConstraints) Validator {
	if s.Normalizer == nil {
		s.Normalizer = PassthroughNormalizer
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimuLinkLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumLinkLength
	}
	return urlValidator(s)
}

func (s urlValidator) Validate(value string) error {
	normalized, err := s.Normalizer.Normalize(value)
	if err != nil {
		return err
	}

	switch length := len(normalized); {
	case length < s.MinimumLength:
		return errors.New("URL is too short")
	case length > s.MaximumLength:
		return errors.New("URL is too long")
	default:
		// _, err := url.Parse(normalized)
		// return err
		return nil
	}
}

type linkTextValidator struct {
	Normalizer    Normalizer
	MinimumLength int
	MaximumLength int
}

func NewLinkTextValidator(s StringConstraints) Validator {
	if s.Normalizer == nil {
		s.Normalizer = PassthroughNormalizer
	}
	if s.MinimumLength < 1 {
		s.MinimumLength = DefaultMinimumLinkTextLength
	}
	if s.MaximumLength < 1 {
		s.MaximumLength = DefaultMaximumLinkTextLength
	}
	return linkTextValidator(s)
}

func (s linkTextValidator) Validate(value string) error {
	normalized, err := s.Normalizer.Normalize(value)
	if err != nil {
		return err
	}
	if normalized != value {
		return errors.New("anchor text is not UTF normalized")
	}

	switch length := len(normalized); {
	case length == 0:
		return errors.New("anchor text is empty")
	case length < s.MinimumLength:
		return errors.New("anchor text is too short")
	case length > s.MaximumLength:
		return errors.New("anchor text is too long")
	default:
		return nil
	}
}

func (r PageValidator) TestLink(origin *url.URL, node *html.Node) func(t *testing.T) {
	return func(t *testing.T) {
		logAttributes(t, node.Attr)
		isEmpty := true
		for descendant := range node.Descendants() {
			switch descendant.Type {
			case html.TextNode:
				// if err := r.LinkText.Validate(ParseTextContent(node)); err != nil {
				// 	t.Errorf("invalid anchor text: %v", err)
				// }
				isEmpty = false
			case html.ElementNode:
				switch strings.ToLower(descendant.Data) {
				case "a", "svg": // contains an image
				}
				isEmpty = false
			}
		}
		if isEmpty {
			t.Error("anchor is empty of meaningful content")
		}

		attributes := getAttributes(t, node)
		rel := make(map[string]struct{})
		if relString, ok := attributes["rel"]; ok {
			for _, directive := range strings.Fields(relString) {
				_, ok = rel[directive]
				if ok {
					t.Errorf("duplicatel a[rel] directive: %s", directive)
				} else {
					rel[directive] = struct{}{}
				}
			}
		}

		if target, ok := attributes["target"]; ok {
			if strings.ToLower(strings.TrimSpace(target)) == "_blank" {
				if _, ok = rel["noopener"]; !ok {
					t.Errorf("anchor text with target=\"_blank\" should have a rel=\"noopener\" setting to prevent tab nabbing; if you need to support older versions of Firefox, use rel=\"noopener noreferrer\"")
				}
			}
		}

		if href, ok := attributes["href"]; ok && len(href) > 0 {
			// TODO: return this as string validator
			// err = r.URL.Validate(href)
			// if err != nil {
			// 	t.Fatalf("dead URL: %v", err)
			// }
			href, _, ok := strings.Cut(href, "#")
			if ok {
				if strings.TrimSpace(href) == "" {
					return // just a #hash reference
				}
			}
			url, err := url.Parse(href)
			if err != nil {
				t.Errorf("failed to parse location: %v", err)
				return
			}
			// TODO: is external should be pulled out of resource
			// path merging should be taken care of when
			// resource is loading
			if IsExternalLocation(origin, url) {
				if _, ok = rel["external"]; !ok {
					t.Error("add \"external\" directive to [rel] attribute")
				}
				if _, ok = rel["nofollow"]; !ok {
					t.Error("add \"nofollow\" directive to [rel] attribute")
				}
			} else {
				url = joinInternalPath(origin, url)
			}
			// _, _, err = r.Loader.Load(t.Context(), href)
			// if err != nil {
			// 	t.Fatalf("unable to load URL <%s>: %v", err, href)
			// }
		} else {
			t.Log("anchor text without href")
		}
	}
}
