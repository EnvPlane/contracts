package domain

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const IssuerMetadataSchemaVersion = "v1"

const (
	IssuerMetadataKindKeys        = "issuer-keys"
	IssuerMetadataKindRevocations = "issuer-revocations"
)

// SignedIssuerMetadata is the public, offline-verifiable transport used for
// issuer key and revocation snapshots. It never carries private key material.
type SignedIssuerMetadata struct {
	Version   string          `json:"version"`
	Kind      string          `json:"kind"`
	IssuedAt  time.Time       `json:"issuedAt"`
	ExpiresAt time.Time       `json:"expiresAt"`
	KeyID     string          `json:"keyId"`
	Algorithm string          `json:"algorithm"`
	Value     json.RawMessage `json:"value"`
	Signature string          `json:"signature"`
}

func IssuerMetadataSigningPayload(version, kind string, issuedAt, expiresAt time.Time, keyID, algorithm string, value any) ([]byte, error) {
	return json.Marshal(struct {
		Version   string    `json:"version"`
		Kind      string    `json:"kind"`
		IssuedAt  time.Time `json:"issuedAt"`
		ExpiresAt time.Time `json:"expiresAt"`
		KeyID     string    `json:"keyId"`
		Algorithm string    `json:"algorithm"`
		Value     any       `json:"value"`
	}{version, kind, issuedAt.UTC(), expiresAt.UTC(), keyID, algorithm, value})
}

func (m SignedIssuerMetadata) Verify(keys map[string]ed25519.PublicKey, expectedKind string, now time.Time) error {
	if m.Version != IssuerMetadataSchemaVersion || m.Kind != expectedKind || m.Algorithm != "Ed25519" || strings.TrimSpace(m.KeyID) == "" || m.IssuedAt.IsZero() || m.ExpiresAt.IsZero() || !m.ExpiresAt.After(m.IssuedAt) || now.Before(m.IssuedAt) || !now.Before(m.ExpiresAt) {
		return errors.New("invalid issuer metadata")
	}
	key, ok := keys[m.KeyID]
	if !ok || len(key) != ed25519.PublicKeySize {
		return errors.New("issuer metadata signing key rejected")
	}
	payload, err := IssuerMetadataSigningPayload(m.Version, m.Kind, m.IssuedAt, m.ExpiresAt, m.KeyID, m.Algorithm, json.RawMessage(m.Value))
	if err != nil {
		return err
	}
	signature, err := base64.RawURLEncoding.DecodeString(m.Signature)
	if err != nil || !ed25519.Verify(key, payload, signature) {
		return errors.New("issuer metadata signature rejected")
	}
	return nil
}
