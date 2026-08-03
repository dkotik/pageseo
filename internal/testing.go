package internal

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/sebdah/goldie/v2"
)

var (
	ErrAtLeastOneInternalTestFailed = errors.New("at least one internal test failed")
	ErrStrangeInternalTestContext   = errors.New("testing: unexpected use of main testing function")
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

func NewTest(name string, run func(*testing.T)) testing.InternalTest {
	return testing.InternalTest{
		Name: newTestName(name),
		F:    run,
	}
}

func NewParallelTest(name string, run func(*testing.T)) testing.InternalTest {
	return testing.InternalTest{
		Name: newTestName(name),
		F: func(t *testing.T) {
			t.Parallel()
			run(t)
		},
	}
}

type testDeps struct{}

func (td testDeps) ModulePath() string                          { return "github.com/dkotik/pageseo" }
func (td testDeps) MatchString(pat, str string) (bool, error)   { return true, nil }
func (td testDeps) StartCPUProfile(w io.Writer) error           { return ErrStrangeInternalTestContext }
func (td testDeps) StopCPUProfile()                             {}
func (td testDeps) WriteProfileTo(string, io.Writer, int) error { return ErrStrangeInternalTestContext }
func (td testDeps) CoordinateFuzzing(time.Duration, int64, time.Duration, int64, int, []corpusEntry, []reflect.Type, string, string) error {
	return nil
}
func (td testDeps) InitRuntimeCoverage() (mode string, tearDown func(coverprofile string, gocoverdir string) (string, error), snapcov func() float64) {
	return "", nil, nil
}
func (td testDeps) RunFuzzWorker(func(corpusEntry) error) error { return ErrStrangeInternalTestContext }
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

func RunTests(set []testing.InternalTest) error {
	switch testing.MainStart(testDeps{}, set, nil, nil, nil).Run() {
	case 0:
		return nil
	default:
		return ErrAtLeastOneInternalTestFailed
	}
}

func RunGoldenTest(t *testing.T, name string, set []testing.InternalTest) {
	// injectedWithFailure := make([]testing.InternalTest, len(set))
	// for i, t := range set {
	// 	runTest := t.F
	// 	injectedWithFailure[i] = NewTest(
	// 		t.Name, func(t *testing.T) {
	// 			runTest(t)
	// 			_, _ = t.Output().Write([]byte("\n\n\n\n"))
	// 			t.Fatal("gold tests always failed to trigger verbose output")
	// 		})
	// }
	if !testing.Verbose() {
		flag.Set("test.v", "true")
		defer flag.Set("test.v", "false")
	}
	joinedOutput := captureOut(func() {
		_ = RunTests(set)
	})
	goldie.New(t).Assert(t, name, joinedOutput)
}

// captureOut captures both stdout and stderr.
func captureOut(f func()) []byte {
	// Create a pipe to capture stdout
	custReader, custWriter, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// Save the original stdout and stderr to restore later
	origStdout := os.Stdout
	origStderr := os.Stderr

	// Restore stdout and stderr when done
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Set the stdout and stderr to the pipe
	os.Stdout, os.Stderr = custWriter, custWriter
	log.SetOutput(custWriter)

	// Create a channel to read the output from the pipe
	out := make(chan []byte)

	// Goroutine reads from pipe and sends output to channel
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		var buf bytes.Buffer
		wg.Done()
		io.Copy(&buf, custReader)
		out <- buf.Bytes()
	}()
	wg.Wait()

	// Call the function that writes to stdout
	f()

	// Close the writer to signal that we're done
	_ = custWriter.Close()

	// Wait for the goroutine to finish reading from the pipe
	return <-out
}
