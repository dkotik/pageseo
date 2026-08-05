# HTML Page Search Engine Optimization Test Suite

[![Go Reference](https://camo.githubusercontent.com/f3bee28c74a644e266e819bedf0150b80af8a7d46292a8fa2837e42aff739ccc/68747470733a2f2f706b672e676f2e6465762f62616467652f6769746875622e636f6d2f5468726565446f74734c6162732f77617465726d696c6c2e737667)](https://pkg.go.dev/github.com/dkotik/pageseo)

Page prevents common page degradation scenarios, which lead to
gradual loss of page ranking:

1. Losing relevant metadata when changing HTML view templates or database models.
2. Duplicating metadata fields on the same page.
3. Duplicating metadata on different publicly available pages.
4. Forgetting to enforce minimum and recommended metadata field sizes.
5. Forgetting to enforce UTF normalization on page content.

The library works by providing a reasonable Goland test suite and a command line tool that any HTML page should pass in order to fit current search engine optimization expectations.

None of the top website fit the "best practice." This indicates that companies are not regularly testing search engine optimization in between hiring consultants.

## Library Usage

When page tests run with `--short` flag, all external page
resources are not loaded or checked.

```sh
go get -u github.com/dkotik/pageseo@latest
```

```go
import (
  "bytes"
  "testing"

  "github.com/dkotik/pageseo"
)

func TestSearchEngineOptimization(t *testing.T) {
  validator := pageseo.NewStrict(
    pageseo.Requirements{
      // override requirements as needed
      Title: pageseo.NewTitleValidator(
        pageseo.StringConstraints{
          MinimumLength: 12,
          MaximumLength: pageseo.DefaultMaximumTitleLength * 4,
          Normalizer: pageseo.NomalizeLineToNFC,
        },
      ),
    },
  )

  t.Run("index.html", validator.TestReader(
    t.Name(), // identify the origin for content de-duplication
    bytes.NewReader([]byte("<html><p>index</p></html>")),
  ))

  t.Run("sitemap.html", validator.TestReader(
    t.Name(), // identify the origin for content de-duplication
    bytes.NewReader([]byte("<html><p>sitemap</p></html>")),
  ))
}
```

## Command Line Usage

### Installation

- MacOS:
  ```sh
  brew tap dkotik/tap
  brew install curl pageseo
  ```
- Debian Package: [latest release](releases)
- Linux Binary: [latest release](releases)
- Windows Binary: [latest release](releases)
- Build from source:
  ```sh
  go install github.com/dkotik/pageseo/cmd/pageseo@latest
  ```

### Scanning

```sh
pageseo --strict --verbose --failfast=false ./**/*.html
```

## Development Road Map

- [x] Provide a command line scanner that can crawl live websites.
- [x] Provide a command line scanner that can validate statically generated websites.
- [x] Unique contraint by namespace with a namespace flag for CLI.
- [ ] Add missing social validators, as image and URL are not tested:
        opengraph.go:55: unknown Open Graph properties:
        opengraph.go:57:  -  og:image:alt
        opengraph.go:57:  -  og:image:height
        opengraph.go:57:  -  og:image:type
        opengraph.go:57:  -  og:image:width
        opengraph.go:57:  -  og:site_name
        opengraph.go:62: og:type not found
        twitter.go:58: unknown Twitter properties:
        twitter.go:60:  -  twitter:image:alt
- [ ] analyze `fb:` meta data
- [ ] support `--json` tag and redirect t.Output() writers to t.Attr()
- [ ] Provide a service that can crawl a target at an interval, and pause at failing crawl until the issue is fixed.
- [ ] Add SQLite backend for the service.
- [ ] Validate image sizes.
- [ ] Validate dependencies in style sheets.

## Similar Projects

- [Front-end Check List](https://github.com/thedaviddias/Front-End-Checklist): an extended page validation list.
- [SEO Crawler](https://github.com/dant89/go-seo): unmaintained.
- [Astro SEO Plugin](https://github.com/jonasmerlin/astro-seo).
