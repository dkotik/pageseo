package main

import (
	"context"
	"errors"
	"flag"
	"math"
	"testing"

	"github.com/urfave/cli/v3"
)

func init() {
	// required to make main function compatible with
	// running [testing.MainStart]
	_ = flag.Uint(flagLimit.Name, flagLimit.Value, flagLimit.Usage)
	_ = flag.Bool(flagFailFast.Name, false, flagFailFast.Usage)
	// _ = flag.Bool(flagStrict.Name, false, flagStrict.Usage)
	_ = flag.Bool(flagShort.Name, false, flagShort.Usage)
	_ = flag.Bool(flagVerbose.Name, false, flagVerbose.Usage)
	testing.Init()
	flag.Parse()
}

var (
	flagLimit = &cli.UintFlag{
		Name:    "limit",
		Aliases: []string{"l"},
		Usage:   "the number of pages at which scanning ends",
		Value:   math.MaxUint16,
		Action: func(_ context.Context, _ *cli.Command, value uint) error {
			if value == 0 {
				return errors.New("limit must be greater than zero")
			}
			return nil
		},
	}

	// flagStrict = &cli.BoolFlag{
	// 	Name:    "strict",
	// 	Aliases: []string{"s"},
	// 	Usage:   "enable strict mode",
	// 	Value:   false,
	// }

	flagCache = &cli.StringFlag{
		Name:    "cache",
		Aliases: []string{"c"},
		Usage:   "cache file for storing page resources",
		Value:   ":memory:",
	}

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

	flagShort = &cli.BoolFlag{
		Name:  "short",
		Usage: "enable short test mode, do not load page resources",
		Value: false,
		Action: func(ctx context.Context, cmd *cli.Command, value bool) error {
			if value {
				flag.Set("test.short", "true")
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
