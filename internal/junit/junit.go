package junit

import (
	"encoding/xml"
	"math"
	"os"
	"strconv"
	"strings"
)

type TypeValue struct {
	Type  string
	Value string
}

type Options struct {
	InputPath      string
	Started        int64
	Tags           []TypeValue
	Fields         []TypeValue
	Release        string
	ReleaseDefault bool
	Suite          string
}

type testSuites struct {
	Suites []testSuite `xml:"testsuite"`
}

type testSuite struct {
	Cases []testCase `xml:"testcase"`
}

type testCase struct {
	ClassName  string       `xml:"classname,attr"`
	Name       string       `xml:"name,attr"`
	Time       string       `xml:"time,attr"`
	Failure    *errorNode   `xml:"failure"`
	Error      *errorNode   `xml:"error"`
	Skipped    *skippedNode `xml:"skipped"`
	Properties *struct {
		Property *struct {
			Value string `xml:"value,attr"`
		} `xml:"property"`
	} `xml:"properties"`
}

type skippedNode struct{}

type errorNode struct {
	Type    string `xml:"type,attr"`
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

type octaneXML struct {
	XMLName     xml.Name      `xml:"test_result"`
	SuiteRef    *idRef        `xml:"suite_ref,omitempty"`
	ReleaseRef  *idRef        `xml:"release_ref,omitempty"`
	Release     *nameRef      `xml:"release,omitempty"`
	TestFields  *testFieldSet `xml:"test_fields,omitempty"`
	Environment *taxonomySet  `xml:"environment,omitempty"`
	TestRuns    testRuns      `xml:"test_runs"`
}

type idRef struct {
	ID string `xml:"id,attr"`
}
type nameRef struct {
	Name string `xml:"name,attr"`
}

type testFieldSet struct {
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
	Type    string `xml:"type,attr,omitempty"`
	Message string `xml:"message,attr,omitempty"`
	Text    string `xml:",chardata"`
}

func BuildInternalResultXML(opts Options) (string, error) {
	raw, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return "", err
	}

	cases, err := readCases(raw)
	if err != nil {
		return "", err
	}

	root := octaneXML{TestRuns: testRuns{Runs: make([]testRun, 0, len(cases))}}
	if opts.Suite != "" {
		root.SuiteRef = &idRef{ID: opts.Suite}
	}
	if opts.ReleaseDefault {
		root.Release = &nameRef{Name: "_default_"}
	} else if opts.Release != "" {
		root.ReleaseRef = &idRef{ID: opts.Release}
	}
	if len(opts.Fields) > 0 {
		set := testFieldSet{}
		for _, f := range opts.Fields {
			set.Fields = append(set.Fields, typedAttr{Type: f.Type, Value: f.Value})
		}
		root.TestFields = &set
	}
	if len(opts.Tags) > 0 {
		set := taxonomySet{}
		for _, t := range opts.Tags {
			set.Entries = append(set.Entries, typedAttr{Type: t.Type, Value: t.Value})
		}
		root.Environment = &set
	}

	for _, tc := range cases {
		pkg, cls := splitClassName(tc.ClassName)
		status := "Passed"
		var errNode *testError
		if tc.Skipped != nil {
			status = "Skipped"
			msg := ""
			if tc.Properties != nil && tc.Properties.Property != nil {
				msg = tc.Properties.Property.Value
			}
			errNode = &testError{Message: msg}
		} else if tc.Failure != nil || tc.Error != nil {
			status = "Failed"
			n := tc.Failure
			if n == nil {
				n = tc.Error
			}
			errNode = &testError{Type: n.Type, Message: n.Message, Text: strings.TrimSpace(n.Text)}
		}
		root.TestRuns.Runs = append(root.TestRuns.Runs, testRun{
			Package:  pkg,
			Class:    cls,
			Name:     choose(tc.Name, "Unnamed test"),
			Status:   status,
			Duration: parseDurationToMs(tc.Time),
			Started:  strconv.FormatInt(opts.Started, 10),
			Error:    errNode,
		})
	}

	bytes, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func readCases(raw []byte) ([]testCase, error) {
	var single testSuite
	if err := xml.Unmarshal(raw, &single); err == nil && len(single.Cases) > 0 {
		return single.Cases, nil
	}
	var suites testSuites
	if err := xml.Unmarshal(raw, &suites); err != nil {
		return nil, err
	}
	var out []testCase
	for _, s := range suites.Suites {
		out = append(out, s.Cases...)
	}
	return out, nil
}

func splitClassName(classname string) (string, string) {
	normalized := strings.TrimSpace(strings.ReplaceAll(classname, "\\", "/"))
	if normalized == "" {
		return "", ""
	}
	if strings.Contains(normalized, "/") {
		parts := strings.Split(normalized, "/")
		filtered := make([]string, 0, len(parts))
		for _, part := range parts {
			if part != "" {
				filtered = append(filtered, part)
			}
		}
		if len(filtered) == 1 {
			return "", filtered[0]
		}
		return strings.Join(filtered[:len(filtered)-1], "."), filtered[len(filtered)-1]
	}
	parts := strings.Split(normalized, ".")
	if len(parts) == 1 {
		return "", parts[0]
	}
	return strings.Join(parts[:len(parts)-1], "."), parts[len(parts)-1]
}

func parseDurationToMs(value string) string {
	if value == "" {
		return "0"
	}
	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil || seconds < 0 {
		return "0"
	}
	return strconv.FormatInt(int64(math.Round(seconds*1000)), 10)
}

func choose(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
