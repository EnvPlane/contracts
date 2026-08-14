package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestAuthenticationSettingsReadModelNeverSerializesClientSecret(t *testing.T) {
	settings := AuthenticationSettings{State: "configured", Mode: "oauth_required", Configured: true}
	payload, err := json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "clientSecret") {
		t.Fatalf("safe read model contains clientSecret: %s", payload)
	}
	command, err := json.Marshal(AuthenticationSettingsCommand{ClientSecret: "write-only"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(command), "clientSecret") {
		t.Fatal("write command must retain write-only credential field")
	}
}
