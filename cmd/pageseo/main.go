package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/dkotik/pageseo"
	"github.com/dkotik/pageseo/crawler"
	"github.com/dkotik/pageseo/crawler/repository"
	"github.com/dkotik/pageseo/internal"
	"github.com/urfave/cli/v3"
	"mvdan.cc/xurls/v2"
	"zombiezen.com/go/sqlite"
)

var (
	atLeastOneInternalTestFailed = false
	errLimitExceeded             = errors.New("limit exceeded")
)

func runTests(set []testing.InternalTest) {
	err := internal.RunTests(set)
	if err == nil {
		return
	}
	if errors.Is(err, internal.ErrAtLeastOneInternalTestFailed) {
		atLeastOneInternalTestFailed = true
		return
	}
	panic(err)
}

func main() {
	cmd := &cli.Command{
		Name:    "pageseo",
		Usage:   "validate HTML page conformity to common search engine optimization practices",
		Version: version(),
		Flags: []cli.Flag{
			flagLimit,
			// flagStrict,
			flagCache,
			flagFailFast,
			flagVerbose,
		},
		Action: cli.ActionFunc(func(ctx context.Context, cmd *cli.Command) (err error) {
			limit := cmd.Uint(flagLimit.Name)
			targets := cmd.Args()
			if !targets.Present() {
				return cli.ShowRootCommandHelp(cmd) // nothing is happening
			}
			// failfast := cmd.Bool(flagFailFast.Name)

			var v pageseo.PageTester
			fsys := os.DirFS(".")
			loader := pageseo.NewFS(fsys)

			v = pageseo.New(loader)
			tests := make([]testing.InternalTest, 0, targets.Len())
			local, remote := separateLocalFromRemoteTargets(targets.Slice())
			if len(local) > 0 {
				for _, target := range local {
					if strings.IndexByte(target, '*') >= 0 {
						matches, err := fs.Glob(fsys, target)
						if err != nil {
							return fmt.Errorf("file path glob failed: %w", err)
						}
						for _, target = range matches {
							if info, err := fs.Stat(fsys, target); err == nil && !info.IsDir() {
								tests = append(tests, internal.NewParallelTest(
									target,
									v.TestFile(target),
								))
							}
						}
					} else {
						tests = append(tests, internal.NewParallelTest(
							target,
							v.TestFile(target),
						))
					}
				}
			}
			total := len(tests)
			if total > 0 {
				tests = tests[:min(total, int(limit))]
				runTests(tests)
				limit = limit - min(limit, uint(len(tests)))
				if limit == 0 {
					return nil
				}
				tests = tests[:0]
			}

			remote = remote[:min(len(remote), int(limit))]
			if len(remote) > 0 {
				conn, err := sqlite.OpenConn(cmd.String(flagCache.Name))
				if err != nil {
					return err
				}
				defer func() {
					err = errors.Join(err, conn.Close())
				}()

				cr, err := crawler.New(
					crawler.AnalyzerFunc(func(ctx context.Context, t repository.Target) error {
						// for _, target := range batch {
						// tests = append(tests, )
						// }
						// runTests(tests)
						// limit = limit - min(limit, uint(len(tests)))
						runTests([]testing.InternalTest{
							internal.NewTest(
								t.Location,
								v.TestPage(t.Location, t.Content),
							),
						})
						limit = limit - 1
						if limit == 0 {
							return errLimitExceeded
						}
						return nil
					}),
					crawler.WithSQLiteConn(conn),
				)
				if err != nil {
					return err
				}

				for _, r := range remote {
					if err = cr.CrawlLocation(ctx, r); err != nil {
						if !errors.Is(err, errLimitExceeded) {
							return err
						}
					}
				}
			}

			if err == nil && total > 0 && !atLeastOneInternalTestFailed {
				fmt.Println(" [🟢] Scanned pages are optimized for search engines.")
			}
			return err
		}),
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf(" [🚫] Unable to analyze pages: %v.\n", err.Error())
	}
}

func version() string {
	v := "dev"
	if info, ok := debug.ReadBuildInfo(); ok {
		v = info.Main.Version
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				v = v + "-" + setting.Value
				break
			}
		}
	}
	return v
}

func separateLocalFromRemoteTargets(targets []string) (local, remote []string) {
nextTarget:
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target == "" {
			continue nextTarget
		}
		url, err := url.Parse(target)
		if err == nil {
			switch strings.ToLower(url.Scheme) {
			case "http", "https":
				remote = append(remote, url.String())
				continue nextTarget
			case "":
				if pageseo.IsLocalHost(url.Hostname()) {
					remote = append(remote, "http://"+target)
					continue nextTarget
				}
				if _, err = os.Stat(target); err != nil {
					if errors.Is(err, os.ErrNotExist) {
						if xurls.Relaxed().MatchString(target) {
							// likely a URL like <truthonly.com> or <www.something...>
							remote = append(remote, "https://"+target)
							continue nextTarget
						}
					}
				}
			}
		}
		local = append(local, url.Path)
	}
	return
}
