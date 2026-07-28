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
			&cli.BoolFlag{
				Name:    "strict",
				Aliases: []string{"s"},
				Usage:   "enable strict mode",
				Value:   false,
			},
			&cli.StringFlag{
				Name:    "namespace",
				Aliases: []string{"n"},
				Usage:   "namespace for metadata unique constraint",
				Value:   "",
			},
			// &cli.BoolFlag{
			// 	Name:    "verbose",
			// 	Aliases: []string{"v"},
			// 	Usage:   "enable verbose output",
			// 	Value:   false,
			// },
		},
		Action: cli.ActionFunc(func(ctx context.Context, cmd *cli.Command) error {
			targets := cmd.Args()
			if !targets.Present() {
				return cli.ShowRootCommandHelp(cmd) // nothing is happening
			}

			var v *pageseo.PageValidator
			if cmd.Bool("strict") {
				v = pageseo.NewStrict(pageseo.Requirements{
					DeduplicationNamespace: cmd.String("namespace"),
				})
			} else {
				v = pageseo.New(pageseo.Requirements{
					DeduplicationNamespace: cmd.String("namespace"),
				})
			}

			tests := make([]testing.InternalTest, 0, targets.Len())
			fsys := os.DirFS(".")
			for _, target := range targets.Slice() {
				if strings.IndexByte(target, '*') >= 0 {
					matches, err := fs.Glob(fsys, target)
					if err != nil {
						return fmt.Errorf("file path glob failed: %w", err)
					}
					for _, target = range matches {
						if info, err := fs.Stat(fsys, target); err == nil && !info.IsDir() {
							tests = append(tests, newTest(ctx, target, v))
						}
					}
				} else {
					tests = append(tests, newTest(ctx, target, v))
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
