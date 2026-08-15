package domain

import (
	"regexp"
	"strings"
	"testing"
)

func TestNormalizeEnvironmentIDProperty(t *testing.T) {
	valid := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	inputs := []string{"", "Feature/ABC_42", "ümlaut branch", strings.Repeat("x", 500), "---", "a..b"}
	for _, input := range inputs {
		got := NormalizeEnvironmentID(input)
		if got != "" && !valid.MatchString(got) {
			t.Fatalf("NormalizeEnvironmentID(%q) = %q", input, got)
		}
		if len(got) > 63 {
			t.Fatalf("NormalizeEnvironmentID(%q) length = %d", input, len(got))
		}
	}
}

func TestEnvironmentWorkBranchIsStableAndModuleIndependent(t *testing.T) {
	got := EnvironmentWorkBranch("Feature/Order_42")
	if got != "envplane/env-feature-order-42" {
		t.Fatalf("work branch = %q", got)
	}
	if strings.Contains(got, ".") || strings.HasPrefix(got, "github.com/") {
		t.Fatalf("work branch contains module or host prefix: %q", got)
	}
}
