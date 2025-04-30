package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/dkotik/pageseo"
	"github.com/urfave/cli/v3"
)

func TestGood(t *testing.T) {
}

func TestBad(t *testing.T) {
	t.Error("This is a mocked failed test")
}

func main() {
	cmd := &cli.Command{
		Name:  "pageseo",
		Usage: "validate HTML page conformity to common search engine optimization practices",
		Commands: []*cli.Command{
			{
				Name:  "scan",
				Usage: "validate HTML page conformity to common search engine optimization practices",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "strict",
						Aliases: []string{"s"},
						Usage:   "enable strict mode",
						Value:   false,
					},
					// &cli.BoolFlag{
					// 	Name:    "verbose",
					// 	Aliases: []string{"v"},
					// 	Usage:   "enable verbose output",
					// 	Value:   false,
					// },
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					targets := cmd.Args()
					if !targets.Present() {
						return fmt.Errorf("no targets provided")
					}

					r := pageseo.Requirements{}
					if cmd.Bool("strict") {
						r = r.WithStrictDefaults()
					} else {
						r = r.WithDefaults()
					}

					tests := make([]testing.InternalTest, 0, targets.Len())
					for _, target := range targets.Slice() {
						tests = append(tests, newTest(ctx, target, r))
					}
					m := testing.MainStart(testDeps{}, tests, nil, nil, nil)
					switch m.Run() {
					case 0:
						fmt.Println("\n🟢 All tests passed.")
						return nil
					default:
						return errors.New("some validation tests failed")
					}
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatalf("🚫 Search engine optimization validation failed: %v.", err)
	}
}
