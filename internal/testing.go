package internal

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
	"reflect"
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

func RunGoldenTest(t *testing.T, goldenFileName, testName string) {
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
	// if !testing.Verbose() {
	// 	flag.Set("test.v", "true")
	// 	defer flag.Set("test.v", "false")
	// }
	// if testing.Short() {
	// 	flag.Set("test.short", "false")
	// 	defer flag.Set("test.short", "true")
	// }
	cmd := exec.Command("go", "test", "-timeout", "5s", "-v", "-run", "^"+testName+"$", ".")
	// joinedOutput := captureStdout(func() {
	// 	_ = RunTests(set)
	// })
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal("failed to run test:", testName, err)
	}
	if len(output) < 4 {
		t.Fatal(testName, "output is too short:", string(output))
	}
	// Last line will contain the result, so we trim it off
	// because the result is dynamic and cannot be predicted.
	//
	// It might contain `(cached)` tag or test duration.
	lastLine := bytes.LastIndex(output[:len(output)-1], []byte("\n"))
	output = output[:lastLine]
	goldie.New(t).Assert(t, goldenFileName, output)
}
