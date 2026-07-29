package main

import (
	"errors"
	"io"
	"net/url"
	"os"
	"reflect"
	"strings"
	"time"

	"mvdan.cc/xurls/v2"
)

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
				if url.Host == "localhost" || url.Host == "127.0.0.1" || url.Host == "::1" {
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
