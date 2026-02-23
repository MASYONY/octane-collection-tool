package playwright

import (
	"encoding/xml"
	"strconv"

	"github.com/example/octane-collection-tool/internal/junit"
)

type TestCase struct {
	PackageName  string
	ClassName    string
	Name         string
	Status       string
	Duration     float64
	ErrorMessage string
	ErrorStack   string
}

type Options struct {
	Started        int64
	Suite          string
	Release        string
	ReleaseDefault bool
	Tags           []junit.TypeValue
	Fields         []junit.TypeValue
	Tests          []TestCase
}

type root struct {
	XMLName     xml.Name     `xml:"test_result"`
	SuiteRef    *idRef       `xml:"suite_ref,omitempty"`
	ReleaseRef  *idRef       `xml:"release_ref,omitempty"`
	Release     *nameRef     `xml:"release,omitempty"`
	TestFields  *fieldSet    `xml:"test_fields,omitempty"`
	Environment *taxonomySet `xml:"environment,omitempty"`
	TestRuns    testRuns     `xml:"test_runs"`
}

type idRef struct {
	ID string `xml:"id,attr"`
}
type nameRef struct {
	Name string `xml:"name,attr"`
}
type fieldSet struct {
	Fields []typedAttr `xml:"test_field"`
}
type taxonomySet struct {
	Entries []typedAttr `xml:"taxonomy"`
}
type typedAttr struct {
	Type  string `xml:"type,attr"`
	Value string `xml:"value,attr"`
}
type testRuns struct {
	Runs []testRun `xml:"test_run"`
}
type testRun struct {
	Package  string     `xml:"package,attr,omitempty"`
	Class    string     `xml:"class,attr,omitempty"`
	Name     string     `xml:"name,attr"`
	Status   string     `xml:"status,attr"`
	Duration string     `xml:"duration,attr"`
	Started  string     `xml:"started,attr"`
	Error    *testError `xml:"error,omitempty"`
}
type testError struct {
	Message string `xml:"message,attr,omitempty"`
	Text    string `xml:",chardata"`
}

func BuildInternalResultXMLFromPlaywright(opts Options) (string, error) {
	root := root{TestRuns: testRuns{Runs: make([]testRun, 0, len(opts.Tests))}}
	if opts.Suite != "" {
		root.SuiteRef = &idRef{ID: opts.Suite}
	}
	if opts.ReleaseDefault {
		root.Release = &nameRef{Name: "_default_"}
	} else if opts.Release != "" {
		root.ReleaseRef = &idRef{ID: opts.Release}
	}
	if len(opts.Fields) > 0 {
		root.TestFields = &fieldSet{}
		for _, f := range opts.Fields {
			root.TestFields.Fields = append(root.TestFields.Fields, typedAttr{Type: f.Type, Value: f.Value})
		}
	}
	if len(opts.Tags) > 0 {
		root.Environment = &taxonomySet{}
		for _, t := range opts.Tags {
			root.Environment.Entries = append(root.Environment.Entries, typedAttr{Type: t.Type, Value: t.Value})
		}
	}
	for _, test := range opts.Tests {
		status := mapStatus(test.Status)
		run := testRun{Package: test.PackageName, Class: test.ClassName, Name: test.Name, Status: status, Duration: strconv.Itoa(int(test.Duration + 0.5)), Started: strconv.FormatInt(opts.Started, 10)}
		if status == "Failed" && (test.ErrorMessage != "" || test.ErrorStack != "") {
			run.Error = &testError{Message: test.ErrorMessage, Text: test.ErrorStack}
		}
		root.TestRuns.Runs = append(root.TestRuns.Runs, run)
	}
	bytes, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func mapStatus(status string) string {
	if status == "passed" {
		return "Passed"
	}
	if status == "skipped" {
		return "Skipped"
	}
	return "Failed"
}
