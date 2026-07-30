package main

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/dkotik/pageseo"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "pageseo",
		Usage:   "validate HTML page conformity to common search engine optimization practices",
		Version: version(),
		Flags: []cli.Flag{
			flagStrict,
			flagFailFast,
			flagVerbose,
		},
		Action: cli.ActionFunc(func(ctx context.Context, cmd *cli.Command) (err error) {
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
								tests = append(tests, newTest(
									target,
									v.TestFile(target),
								))
							}
						}
					} else {
						tests = append(tests, newTest(
							target,
							v.TestFile(target),
						))
					}
				}
			}
			runTests(tests)
			tests = tests[:0]

			if len(remote) > 0 {
				queue := newResourceQueue(remote...)
				loader := pageseo.NewCache(queue.Push).WrapLoader(newClientPool(6))
				v.Loader = loader
				for _, target := range remote {
					_, _, _ = loader.Load(ctx, target)
				}

				for {
					batch := queue.Pull()
					if len(batch) == 0 {
						return nil // all finished
					}
					for _, target := range batch {
						tests = append(tests, newTest(
							target.URL,
							v.TestReader(target.URL, bytes.NewReader(target.Content)),
						))
					}
					runTests(tests)
				}
			}
			return nil
		}),
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Printf(" [🚫] Unable to analyze pages: %v.\n", err.Error())
	} else if allTestsArePassing {
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
