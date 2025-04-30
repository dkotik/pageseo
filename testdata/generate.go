package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var sources = [...]string{
	// "https://google.com",
	"https://wikipedia.org",
	// "https://youtube.com",
	"https://microsoft.com",
	// "https://apple.com",
	"https://amazon.com",
	// "https://yahoo.com",
	"https://dw.com",
	"https://bbc.com",
	"https://cnn.com",
	// "https://nytimes.com",
	"https://www.theguardian.com/europe",
}

func fileName(s string) string {
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "www.")
	s, _, _ = strings.Cut(s, ".")
	return fmt.Sprintf("./testdata/popular/%s.html", s)
}

func download(source, destination string) error {
	r, err := http.Get(source)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	f, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, r.Body)
	return err
}

func main() {
	var err error
	for _, source := range sources {
		if err = download(source, fileName(source)); err != nil {
			panic(err)
		}
	}
}
