package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

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
				Action: func(ctx context.Context, cmd *cli.Command) error {
					targets := cmd.Args()
					if !targets.Present() {
						return fmt.Errorf("no targets provided")
					}
					tests := make([]testing.InternalTest, 0, targets.Len())
					for _, target := range targets.Slice() {
						tests = append(tests, newTest(target))
					}
					m := testing.MainStart(testDeps{}, tests, nil, nil, nil)
					switch m.Run() {
					case 0:
						fmt.Println("\n🟢 All tests passed.")
					default:
						fmt.Println("\n🚫 Problems found.")
					}
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
