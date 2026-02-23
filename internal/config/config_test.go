package config

import "testing"

func TestParseProperties(t *testing.T) {
	parsed := ParseProperties("#comment\nserver=http://localhost:8080\nsharedspace=1001\nworkspace=1002")
	if parsed["server"] != "http://localhost:8080" || parsed["sharedspace"] != "1001" || parsed["workspace"] != "1002" {
		t.Fatalf("unexpected parse result: %#v", parsed)
	}
}
