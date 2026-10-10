package contracts

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestContractsVersionGuardDiscoversAllDirectConsumers(t *testing.T) {
	root := t.TempDir()
	write := func(relative, content string) {
		t.Helper()
		path := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("contracts/go.mod", "module github.com/envplane/contracts\n")
	run := func(expected string) (string, error) {
		t.Helper()
		cmd := exec.Command("bash", "scripts/check-contract-version.sh", root) // #nosec G204 -- fixed checked-in script; root is a test-created fixture path, not executable content.
		for _, v := range os.Environ() {
			if !strings.HasPrefix(v, "ENVPLANE_CONTRACTS_VERSION=") {
				cmd.Env = append(cmd.Env, v)
			}
		}
		if expected != "" {
			cmd.Env = append(cmd.Env, "ENVPLANE_CONTRACTS_VERSION="+expected)
		}
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	if out, err := run(""); err != nil {
		t.Fatalf("standalone contracts check failed: %s %v", out, err)
	}
	write("activation-issuer/go.mod", "module activation\nrequire github.com/envplane/contracts v0.0.1\n")
	write("PRIVATE/nested-consumer/go.mod", "module nested\nrequire (\n github.com/envplane/contracts v0.0.1\n)\n")
	if out, err := run(""); err == nil || !strings.Contains(out, "verified published tag") {
		t.Fatal("guard guessed a version instead of requiring verified release context")
	}
	if out, err := run("v0.0.1"); err != nil {
		t.Fatalf("aligned consumers rejected: %s %v", out, err)
	}
	write("PRIVATE/nested-consumer/go.mod", "module nested\nrequire github.com/envplane/contracts v0.0.2\n")
	if out, err := run("v0.0.1"); err == nil || !strings.Contains(out, "PRIVATE/nested-consumer/go.mod") {
		t.Fatal("nested PRIVATE consumer not checked")
	}
}
