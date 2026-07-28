package main

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dkotik/pageseo"
	"mvdan.cc/xurls/v2"
)

func newTestName(target string) string {
	i := len(target) - 1
	for ; i >= 0; i-- {
		switch target[i] {
		case '\\', '/':
			continue
		default:
			return target[:i+1]
		}
	}
	return target
}

func newTest(ctx context.Context, target string, r *pageseo.PageValidator) testing.InternalTest {
	target = strings.TrimSpace(target)
	if target == "" {
		return testing.InternalTest{
			Name: newTestName(target),
			F: func(t *testing.T) {
				t.Fatal("invalid test target:", target)
			},
		}
	}

	url, err := url.Parse(target)
	if err == nil {
		switch strings.ToLower(url.Scheme) {
		case "http", "https":
			return testing.InternalTest{
				Name: newTestName(target),
				F:    r.TestURL(ctx, url.String()),
			}
		case "file":
			target = url.Path
		case "":
			if _, err = os.Stat(target); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if xurls.Relaxed().MatchString(target) {
						// likely a URL like <truthonly.com> or <www.something...>
						return testing.InternalTest{
							Name: newTestName(target),
							F:    r.TestURL(ctx, "https://"+target),
						}
					}
				} else {
					return testing.InternalTest{
						Name: newTestName(target),
						F: func(t *testing.T) {
							t.Fatal("invalid test target:", target)
						},
					}
				}
			}
		}
	}

	return testing.InternalTest{
		Name: newTestName(target),
		F:    r.TestFile(target),
	}
}

type testDeps struct{}

func (td testDeps) ModulePath() string                          { return "github.com/dkotik/pageseo" }
func (td testDeps) MatchString(pat, str string) (bool, error)   { return true, nil }
func (td testDeps) StartCPUProfile(w io.Writer) error           { return nil }
func (td testDeps) StopCPUProfile()                             {}
func (td testDeps) WriteProfileTo(string, io.Writer, int) error { return nil }
func (td testDeps) CoordinateFuzzing(time.Duration, int64, time.Duration, int64, int, []corpusEntry, []reflect.Type, string, string) error {
	return nil
}
func (td testDeps) InitRuntimeCoverage() (mode string, tearDown func(coverprofile string, gocoverdir string) (string, error), snapcov func() float64) {
	return "", nil, nil
}
func (td testDeps) RunFuzzWorker(func(corpusEntry) error) error { return nil }
func (td testDeps) ReadCorpus(string, []reflect.Type) ([]corpusEntry, error) {
	return nil, nil
}
func (td testDeps) ResetCoverage()                          {}
func (td testDeps) ImportPath() string                      { return "" }
func (td testDeps) StartTestLog(io.Writer)                  {}
func (td testDeps) StopTestLog() error                      { return nil }
func (td testDeps) SetPanicOnExit0(bool)                    {}
func (td testDeps) SnapshotCoverage()                       {}
func (td testDeps) CheckCorpus([]any, []reflect.Type) error { return nil }

type corpusEntry = struct {
	Parent     string
	Path       string
	Data       []byte
	Values     []any
	Generation int
	IsSeed     bool
}
