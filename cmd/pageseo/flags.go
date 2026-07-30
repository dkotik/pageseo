package main

import (
	"context"
	"flag"
	"testing"

	"github.com/urfave/cli/v3"
)

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
	flagVerbose = &cli.BoolFlag{
		Name: "verbose",
		// Aliases: []string{"v", "test.v"},
		Usage: "enable verbose output",
		Value: false,
		Action: func(ctx context.Context, cmd *cli.Command, value bool) error {
			if value {
				_ = flag.Bool("verbose", false, "a bool flag")
				testing.Init()
				flag.Set("test.v", "true")
				flag.Parse()
			}
			return nil
		},
	}
)
