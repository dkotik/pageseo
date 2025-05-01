package htmltest

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"slices"
)

var reExternalURL = regexp.MustCompile(`^(\w+:)?/?/`)

func IsLocalURL(s string) bool {
	return !reExternalURL.MatchString(s)
}

func JoinURL(base, path string) (string, error) {
	// reExternalURL.FindStringIndex(path)
	if IsLocalURL(path) {
		return url.JoinPath(base, path)
	}
	return path, nil
}

type urlValidator struct {
	StatusCodes []int
	Client      *http.Client
}

func NewURLValidator(client *http.Client, acceptStatusCodes ...int) Validator {
	if client == nil {
		client = http.DefaultClient
	}
	if len(acceptStatusCodes) == 0 {
		acceptStatusCodes = []int{http.StatusOK}
	}
	return &urlValidator{
		Client:      client,
		StatusCodes: acceptStatusCodes,
	}
}

func (v *urlValidator) Validate(s string) error {
	parsed, err := url.Parse(s)
	if err != nil {
		return err
	}
	switch parsed.Scheme {
	case "file":
		file, err := os.Open(s)
		if err != nil {
			return err
		}
		defer file.Close()
		return nil
	case "http", "https":
		response, err := v.Client.Get(s)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if !slices.Contains(v.StatusCodes, response.StatusCode) {
			return fmt.Errorf("unexpected response status code: %d - %s", response.StatusCode, http.StatusText(response.StatusCode))
		}
		return nil
	case "":
		return nil
	default:
		return fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
	}
}
