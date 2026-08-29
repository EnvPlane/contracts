package domain

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestSignedIssuerMetadataRejectsTamperingWrongKeyAndExpiry(t *testing.T) {
	seed := sha256.Sum256([]byte("issuer-metadata-v1"))
	private := ed25519.NewKeyFromSeed(seed[:])
	now := time.Date(2026, 8, 29, 0, 0, 0, 0, time.UTC)
	value, err := json.Marshal(IssuerRevocationMetadata{Revocations: []IssuerLicenseRevocation{{LicenseID: "license-a", RevokedAt: now}}})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := IssuerMetadataSigningPayload(IssuerMetadataSchemaVersion, IssuerMetadataKindRevocations, now, now.Add(time.Hour), "issuer-1", "Ed25519", json.RawMessage(value))
	if err != nil {
		t.Fatal(err)
	}
	metadata := SignedIssuerMetadata{Version: IssuerMetadataSchemaVersion, Kind: IssuerMetadataKindRevocations, IssuedAt: now, ExpiresAt: now.Add(time.Hour), KeyID: "issuer-1", Algorithm: "Ed25519", Value: value, Signature: base64.RawURLEncoding.EncodeToString(ed25519.Sign(private, payload))}
	keys := map[string]ed25519.PublicKey{"issuer-1": private.Public().(ed25519.PublicKey)}
	if err := metadata.Verify(keys, IssuerMetadataKindRevocations, now); err != nil {
		t.Fatal(err)
	}
	metadata.Value = json.RawMessage(`[]`)
	if err := metadata.Verify(keys, IssuerMetadataKindRevocations, now); err == nil {
		t.Fatal("tampering accepted")
	}
	metadata.Value = value
	if err := metadata.Verify(map[string]ed25519.PublicKey{}, IssuerMetadataKindRevocations, now); err == nil {
		t.Fatal("unknown key accepted")
	}
	if err := metadata.Verify(keys, IssuerMetadataKindRevocations, now.Add(2*time.Hour)); err == nil {
		t.Fatal("expired metadata accepted")
	}
}
