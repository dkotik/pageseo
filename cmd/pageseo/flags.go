package main

import (
	"context"
	"flag"
	"testing"

	"github.com/urfave/cli/v3"
)

func init() {
	// required to make main function compatible with
	// running [testing.MainStart]
	_ = flag.Bool("failfast", false, flagFailFast.Usage)
	_ = flag.Bool("verbose", false, flagVerbose.Usage)
	testing.Init()
	flag.Parse()
}

var (
	flagStrict = &cli.BoolFlag{
		Name:    "strict",
		Aliases: []string{"s"},
		Usage:   "enable strict mode",
		Value:   false,
	}

	//	flagNameSpace =			&cli.StringFlag{
	//					Name:    "namespace",
	//					Aliases: []string{"n"},
	//					Usage:   "namespace for metadata unique constraint",
	//					Value:   "",
	//				}

	flagFailFast = &cli.BoolFlag{
		Name:  "failfast",
		Usage: "end the test on the first detected failure",
		Value: false,
		Action: func(ctx context.Context, cmd *cli.Command, value bool) error {
			if value {
				flag.Set("test.failfast", "true")
			}
			return nil
		},
	}

	flagVerbose = &cli.BoolFlag{
		Name: "verbose",
		// Aliases: []string{"v", "test.v"},
		Usage: "enable verbose output",
		Value: false,
		Action: func(ctx context.Context, cmd *cli.Command, value bool) error {
			if value {
				flag.Set("test.v", "true")
			}
			return nil
		},
	}
)
