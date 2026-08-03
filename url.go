package pageseo

import (
	"errors"
	"net/url"
	"path"
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

func IsSubdomainOfOrigin(origin, location *url.URL) bool {
	root := strings.SplitN(origin.Hostname(), ".", 6)
	switch len(root) {
	case 0:
		return false
	case 1:
		return strings.HasSuffix(location.Host, "."+root[0])
	default:
		return strings.HasSuffix(location.Host, "."+strings.Join(root[1:], "."))
	}
}

func JoinPath(origin, location *url.URL) (result *url.URL, isExternal bool) {
	if IsExternalLocation(origin, location) {
		return location, true
	}
	return joinInternalPath(origin, location), false
}

func joinRelativePath(origin *url.URL, location string) string {
	parsed, err := url.Parse(location)
	if err != nil {
		return location // unparseable
	}
	if IsExternalLocation(origin, parsed) {
		parsed.Path = path.Clean(parsed.Path)
		return parsed.String()
	}
	return joinInternalPath(origin, parsed).String()
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
			Path:        path.Clean(location.Path),
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
