package domain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"testing"
)

func TestStorageEvidenceSigningDomainsCannotBeSubstituted(t *testing.T) {
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(StorageErasureStatement{Version: "1", Coverage: "volume_only", Method: "cryptographic_erasure"})
	if err != nil {
		t.Fatal(err)
	}
	report := append([]byte(StorageProviderReportSigningDomain), data...)
	erasure := append([]byte(StorageErasureStatementSigningDomain), data...)
	signature := ed25519.Sign(private, report)
	if !ed25519.Verify(public, report, signature) || ed25519.Verify(public, erasure, signature) {
		t.Fatal("deletion receipt substituted for erasure certificate")
	}
	if !bytes.Contains(data, []byte(`"keyDestroyed":false`)) || !bytes.Contains(data, []byte(`"keyExclusive":false`)) || !bytes.Contains(data, []byte(`"keyCopiesDestroyed":false`)) {
		t.Fatal("missing explicit cryptographic evidence requirements")
	}
}
