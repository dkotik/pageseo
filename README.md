# HTML Page Search Engine Optimization

HTML page search engine optimization test and utility suite. Aims to prevent the following common page degradation scenarios, which lead to loss of page ranking:

1. Losing relevant metadata when changing HTML view templates or database models.
2. Duplicating metadata fields on the same page.
3. Duplicating metadata on different publicly available pages.
4. Forgetting to enforce minimum and recommended metadata field sizes.
5. Forgetting to enforce UTF normalization on page content.

The library works by providing a reasonable set of tests that any HTML must conform to in order to fit the best current search engine optimization expectations.

## Library Usage

```sh
go get -u github.com/dkotik/pageseo@latest
```

```go
import (
	"bytes"
	"testing"

	"github.com/dkotik/pageseo"
)

func TestViewSearchEngineOptimization(t *testing.T) {
	view := "<html></html>"

	pageseo.Requirements{}.WithStrictDefaults().TestReader(
		bytes.NewReader(view),
	)
}

```

## Command Line Usage

```sh
go install github.com/dkotik/pageseo/cmd/pageseo@latest
pageseo scan --strict ./**/*.html
```

## Development Road Map

Project status: draft in progress.

- [x] Provide a command line scanner that can validate statically generated websites.
- [ ] Provide a command line scanner that can crawl live websites.
- [ ] Unique contraint by namespace with a namespace flag for CLI.
- [ ] Add open graph validations.

## Similar Projects

- [Front-end Check List](https://github.com/thedaviddias/Front-End-Checklist): an extended page validation list.
- [SEO Crawler](https://github.com/dant89/go-seo): unmaintained.
- [Astro SEO Plugin](https://github.com/jonasmerlin/astro-seo).
