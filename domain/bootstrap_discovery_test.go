package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBootstrapAgentStatusSerializesEmptyResourceCount(t *testing.T) {
	payload, err := json.Marshal(BootstrapAgentStatusResponse{
		Status:             "online",
		ResourceScanStatus: "completed",
		ResourceCount:      0,
	})
	if err != nil {
		t.Fatalf("marshal bootstrap agent status: %v", err)
	}
	if !strings.Contains(string(payload), `"resourceCount":0`) {
		t.Fatalf("completed empty scan must serialize resourceCount: %s", payload)
	}
}
