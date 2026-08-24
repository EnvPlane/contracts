package domain

import (
	"strings"
	"testing"
	"time"
)

func sandboxSpec() AISandboxSpec {
	return AISandboxSpec{SchemaVersion: AISandboxSchemaVersion, RunID: "run-1", TenantID: "tenant-a", ProjectID: "project-a", Transform: AISandboxCompareState, Image: "registry.example/analysis@sha256:" + strings.Repeat("a", 64), Inputs: []AISignedInputRef{{ID: "input-1", Digest: "sha256:" + strings.Repeat("b", 64), Signature: "sig", KeyID: "key-1"}}, NetworkDenied: true, Resources: AISandboxResourceLimits{CPUMillis: 500, MemoryMiB: 256, PIDs: 32, TimeoutSeconds: 60, MaxOutputBytes: 64 * 1024}, CreatedAt: time.Now().UTC()}
}

func TestAISandboxRejectsShellNetworkAndMutableImages(t *testing.T) {
	spec := sandboxSpec()
	if err := spec.Validate(); err != nil {
		t.Fatal(err)
	}
	spec.Image = "registry.example/analysis:latest"
	if err := spec.Validate(); err == nil {
		t.Fatal("mutable image accepted")
	}
	spec = sandboxSpec()
	spec.NetworkDenied = false
	spec.Egress = []AISandboxEgress{{Host: "0.0.0.0/0", Port: 443}}
	if err := spec.Validate(); err == nil {
		t.Fatal("implicit network access accepted")
	}
}

func TestAISandboxOutputIsBounded(t *testing.T) {
	output := AISandboxOutput{SchemaVersion: AISandboxSchemaVersion, RunID: "run-1", Digest: "sha256:out", Signature: "sig", KeyID: "key-1", Payload: "{}"}
	if err := output.Validate(64); err != nil {
		t.Fatal(err)
	}
	output.Payload = strings.Repeat("x", 65)
	if err := output.Validate(64); err == nil {
		t.Fatal("oversized output accepted")
	}
}
