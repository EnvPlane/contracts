package domain

import "testing"

func TestParseJobStatusContract(t *testing.T) {
	for _, status := range AllJobStatuses() {
		if got, err := ParseJobStatus(string(status)); err != nil || got != status {
			t.Fatalf("status %q: got %q, err=%v", status, got, err)
		}
	}
	if _, err := ParseJobStatus("unknown"); err == nil {
		t.Fatal("expected unknown status to be rejected")
	}
}
