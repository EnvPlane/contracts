package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSCMWebhookBootstrapProofReady(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint, receiver, delivery string
		want                               bool
	}{
		{"ready", "ready", "ready", "verified", true},
		{"endpoint pending", "pending", "ready", "verified", false},
		{"receiver unknown", "ready", "unknown", "verified", false},
		{"delivery pending", "ready", "ready", "pending", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := (SCMWebhookBootstrapProof{EndpointState: tc.endpoint, ReceiverState: tc.receiver, DeliveryState: tc.delivery}).Ready(); got != tc.want {
				t.Fatalf("Ready()=%v", got)
			}
		})
	}
}

func TestSCMWebhookBootstrapProofReadyForConfig(t *testing.T) {
	proof := SCMWebhookBootstrapProof{EndpointState: "ready", ReceiverState: "ready", DeliveryState: "verified", ConfigFingerprint: "current"}
	for _, tc := range []struct {
		name, expected string
		want           bool
	}{
		{name: "matching configuration", expected: "current", want: true},
		{name: "stale configuration", expected: "changed", want: false},
		{name: "missing expected configuration", expected: "", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := proof.ReadyForConfig(tc.expected); got != tc.want {
				t.Fatalf("ReadyForConfig(%q) = %v, want %v", tc.expected, got, tc.want)
			}
		})
	}
}

func TestSCMWebhookBootstrapProofJSONContainsNoCredentialFields(t *testing.T) {
	b, err := json.Marshal(SCMWebhookBootstrapProof{SecretFingerprint: "sha256:abc"})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"token", "secret"} {
		if strings.Contains(strings.ToLower(string(b)), `"`+forbidden+`"`) {
			t.Fatalf("credential field %q leaked: %s", forbidden, b)
		}
	}
}
