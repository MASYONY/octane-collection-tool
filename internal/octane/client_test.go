package octane

import "testing"

func TestParseResultLog(t *testing.T) {
	parsed := ParseResultLog("status: failed\nuntil: 2026-02-10T10:00:00Z\nError Msg")
	if parsed.Status != "failed" || parsed.Until != "2026-02-10T10:00:00Z" || parsed.ErrorDetails != "Error Msg" {
		t.Fatalf("unexpected parse result: %#v", parsed)
	}
}
