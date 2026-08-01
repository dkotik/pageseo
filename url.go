package pageseo

import (
	"errors"
	"net/url"
	"strings"
	"sync"
)

type cachedParsedURLs struct {
	URLs sync.Map
}

func (c *cachedParsedURLs) GetRelativeTo(origin *url.URL, urlStr string) (location *url.URL, isExternal bool, err error) {
	location, err = c.Get(urlStr)
	if err != nil {
		return nil, false, err
	}
	location, isExternal = JoinPath(origin, location)
	return location, isExternal, nil
}

func (c *cachedParsedURLs) Get(urlStr string) (*url.URL, error) {
	if urlStr == "" {
		return nil, errors.New("URL is empty")
	}
	if cached, ok := c.URLs.Load(urlStr); ok {
		return cached.(*url.URL), nil
	}
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	c.URLs.Store(urlStr, parsed)
	return parsed, nil
}

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

func JoinPath(origin, location *url.URL) (result *url.URL, isExternal bool) {
	if IsExternalLocation(origin, location) {
		return location, true
	}
	return joinInternalPath(origin, location), false
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
		if index := strings.IndexByte(normalized, '?'); index != -1 && index < s.MaximumLength {
			return nil // disregard the query string when weighing length
		}
		return errors.New("URL is too long")
	default:
		// _, err := url.Parse(normalized)
		// return err
		return nil
	}
}
