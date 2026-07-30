package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
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
			flagVerbose,
		},
		Action: cli.ActionFunc(func(ctx context.Context, cmd *cli.Command) error {
			targets := cmd.Args()
			if !targets.Present() {
				return cli.ShowRootCommandHelp(cmd) // nothing is happening
			}

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
			if len(remote) > 0 {
				crawler := pageseo.NewCrawler(
					v,
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
					newClientHTTP(),
				)
				for _, target := range remote {
					tests = append(tests, newTest(
						target,
						crawler(target),
					))
				}
			}

			m := testing.MainStart(testDeps{}, tests, nil, nil, nil)
			switch m.Run() {
			case 0:
				fmt.Println("\n🟢 All tests passed.")
				return nil
			default:
				return errors.New("some validation tests failed")
			}
		}),
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("🚫 Search engine optimization validation failed: %v.", err)
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
