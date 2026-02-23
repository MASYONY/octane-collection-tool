package junit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildInternalResultXML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "junit.xml")
	_ = os.WriteFile(p, []byte(`<testsuite><testcase classname="com.example.A" name="ok" time="0.1"/><testcase classname="x.y.B" name="bad"><failure message="oops" type="AssertionError">trace</failure></testcase></testsuite>`), 0644)
	xml, err := BuildInternalResultXML(Options{InputPath: p, Started: 10, Tags: []TypeValue{{Type: "OS", Value: "Linux"}}, Fields: []TypeValue{{Type: "Framework", Value: "JUnit"}}, Suite: "100", Release: "200"})
	if err != nil {
		t.Fatal(err)
	}
	for _, expect := range []string{"<suite_ref id=\"100\"></suite_ref>", "<release_ref id=\"200\"></release_ref>", "status=\"Passed\"", "status=\"Failed\"", "<error type=\"AssertionError\" message=\"oops\">trace</error>"} {
		if !strings.Contains(xml, expect) {
			t.Fatalf("missing %q in xml: %s", expect, xml)
		}
	}
}
