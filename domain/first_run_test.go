package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFirstRunProgressTransitionsAreSequentialAndIdempotent(t *testing.T) {
	progress := InitialFirstRunProgress()
	if !progress.CanAdvanceTo(FirstRunUninitialized) || !progress.CanAdvanceTo(FirstRunOperatorClaimed) {
		t.Fatal("initial progress must permit idempotent claim and its next state")
	}
	if progress.CanAdvanceTo(FirstRunClusterReady) {
		t.Fatal("first-run progress must reject skipped transitions")
	}
}

func TestFirstRunProgressJSONContainsNoCredentialFields(t *testing.T) {
	data, err := json.Marshal(InitialFirstRunProgress())
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(data))
	for _, forbidden := range []string{"token", "secret", "password", "claimedby", "email"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("first-run progress serialized unsafe field %q: %s", forbidden, data)
		}
	}
}
