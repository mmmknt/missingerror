package missingerror_test

import (
	"testing"

	"github.com/gostaticanalysis/testutil"
	"github.com/mmmknt/missingerror"
	"golang.org/x/tools/go/analysis/analysistest"
)

func init() {
	missingerror.Analyzer.Flags.Set("wrappers", "fmt.Errorf,a/helper.Wrap")
}

// TestAnalyzer is a test for Analyzer.
func TestAnalyzer(t *testing.T) {
	testdata := testutil.WithModules(t, analysistest.TestData(), nil)
	analysistest.Run(t, testdata, missingerror.Analyzer, "a")
}
