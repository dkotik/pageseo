# HTML Page Search Engine Optimization Test Suite

HTML page search engine optimization test and utility Golang test suite. Aims to prevent the following common page degradation scenarios, which lead to loss of page ranking:

1. Losing relevant metadata when changing HTML view templates or database models.
2. Duplicating metadata fields on the same page.
3. Duplicating metadata on different publicly available pages.
4. Forgetting to enforce minimum and recommended metadata field sizes.
5. Forgetting to enforce UTF normalization on page content.

The library works by providing a reasonable set of tests that any HTML page should pass in order to fit current search engine optimization expectations. Almost none of the top website fit the "best practice." This indicates that almost nobody is testing search engine optimization in between hiring consultants. The library aims to be minimal and flexible to your use case.

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

```sh
go install github.com/dkotik/pageseo/cmd/pageseo@latest
pageseo scan --strict ./**/*.html
```

## Development Road Map

**Project status: draft in progress. The test suite is minimal, but in strict mode it will find at least one reasonable optimization suggestion for your website.**

- [x] Provide a command line scanner that can validate statically generated websites.
- [x] Provide a command line scanner that can validate URLs.
- [x] Add open graph validations.
- [x] Add twitter validations.
- [x] Unique contraint by namespace with a namespace flag for CLI.
- [ ] Validate image size.
- [ ] Provide a command line scanner that can crawl live websites.
- [ ] Provide a service that can crawl a target at an interval.
- [ ] Make sure `--failfast` works for CLI.

## Similar Projects

- [Front-end Check List](https://github.com/thedaviddias/Front-End-Checklist): an extended page validation list.
- [SEO Crawler](https://github.com/dant89/go-seo): unmaintained.
- [Astro SEO Plugin](https://github.com/jonasmerlin/astro-seo).
