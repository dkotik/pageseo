package main

import (
	"bytes"
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
	"github.com/dkotik/pageseo/internal"
	"github.com/urfave/cli/v3"
	"mvdan.cc/xurls/v2"
)

var atLeastOneInternalTestFailed = false

func runTests(set []testing.InternalTest) {
	err := internal.RunTests(set)
	if err == nil {
		return
	}
	if errors.Is(err, internal.ErrAtLeastOneInternalTestFailed) {
		atLeastOneInternalTestFailed = true
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
			flagStrict,
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

			var v pageseo.PageValidator
			fsys := os.DirFS(".")
			loader := pageseo.NewFS(fsys)
			if cmd.Bool("strict") {
				v = pageseo.NewStrict(loader, pageseo.Requirements{
					DeduplicationNamespace: cmd.String("namespace"),
				})
			} else {
				v = pageseo.New(loader, pageseo.Requirements{
					DeduplicationNamespace: cmd.String("namespace"),
				})
			}
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
			if total := len(tests); total > 0 {
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
				queue := newResourceQueue(remote...)
				loader := pageseo.NewCache(queue.Push).WrapLoader(newClientPool(6))
				v.Loader = loader
				for _, target := range remote {
					_, _, _ = loader.Load(ctx, target)
				}

				for {
					batch := queue.Pull()
					batch = batch[:min(len(batch), int(limit))]
					if len(batch) == 0 {
						return nil // all finished
					}
					for _, target := range batch {
						tests = append(tests, internal.NewParallelTest(
							target.URL,
							v.TestReader(target.URL, bytes.NewReader(target.Content)),
						))
					}
					runTests(tests)
					limit = limit - min(limit, uint(len(tests)))
				}
			}
			return nil
		}),
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf(" [🚫] Unable to analyze pages: %v.\n", err.Error())
	} else if !atLeastOneInternalTestFailed {
		fmt.Println(" [🟢] Scanned pages are optimized for search engines.")
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
