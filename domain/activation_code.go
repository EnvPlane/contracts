package domain

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const ActivationCodeSchemaVersion = "v1"
const ActivationCodeMaxLength = 8192

type CommercialSnapshot struct {
	Currency        string `json:"currency"`
	AmountMinor     int64  `json:"amountMinor"`
	BillingInterval string `json:"billingInterval"`
	TaxMode         string `json:"taxMode"`
}

type ActivationGrant struct {
	SchemaVersion  string             `json:"schemaVersion"`
	InstallationID string             `json:"installationId"`
	TenantID       string             `json:"tenantId"`
	SKU            string             `json:"sku"`
	PlanID         string             `json:"planId"`
	PlanVersion    string             `json:"planVersion"`
	Features       map[string]bool    `json:"features"`
	Limits         map[string]int64   `json:"limits"`
	IssuedAt       time.Time          `json:"issuedAt"`
	NotBefore      time.Time          `json:"notBefore"`
	ExpiresAt      time.Time          `json:"expiresAt"`
	LicenseID      string             `json:"licenseId"`
	Nonce          string             `json:"nonce"`
	Commercial     CommercialSnapshot `json:"commercial"`
}

type ActivationCodeBinding struct {
	InstallationID string
	TenantID       string
}
type activationCodeEnvelope struct {
	Version   string          `json:"v"`
	KeyID     string          `json:"kid"`
	Algorithm string          `json:"alg"`
	Grant     ActivationGrant `json:"grant"`
}

func (g ActivationGrant) Validate(now time.Time, binding ActivationCodeBinding) error {
	if g.SchemaVersion != ActivationCodeSchemaVersion || strings.TrimSpace(g.InstallationID) == "" || strings.TrimSpace(g.TenantID) == "" || strings.TrimSpace(g.SKU) == "" || strings.TrimSpace(g.PlanID) == "" || strings.TrimSpace(g.PlanVersion) == "" || strings.TrimSpace(g.LicenseID) == "" || strings.TrimSpace(g.Nonce) == "" {
		return errors.New("invalid activation grant")
	}
	if g.InstallationID != binding.InstallationID || g.TenantID != binding.TenantID {
		return errors.New("activation grant binding mismatch")
	}
	if g.IssuedAt.IsZero() || g.NotBefore.IsZero() || g.ExpiresAt.IsZero() || !g.ExpiresAt.After(g.NotBefore) || now.Before(g.NotBefore) || !now.Before(g.ExpiresAt) {
		return errors.New("activation grant is not currently valid")
	}
	if g.Commercial.AmountMinor < 0 || strings.TrimSpace(g.Commercial.Currency) == "" || strings.TrimSpace(g.Commercial.BillingInterval) == "" || strings.TrimSpace(g.Commercial.TaxMode) == "" {
		return errors.New("invalid commercial snapshot")
	}
	for key, value := range g.Limits {
		if strings.TrimSpace(key) == "" || value < 0 {
			return errors.New("invalid activation limit")
		}
	}
	return nil
}

// Authorized intentionally ignores Commercial: price is signed display data,
// never an authorization input.
func (g ActivationGrant) Authorized(feature string, limit string) (bool, int64) {
	return g.Features[feature], g.Limits[limit]
}

func SignActivationCode(grant ActivationGrant, keyID string, privateKey ed25519.PrivateKey) (string, error) {
	if err := grant.Validate(grant.IssuedAt, ActivationCodeBinding{InstallationID: grant.InstallationID, TenantID: grant.TenantID}); err != nil || strings.TrimSpace(keyID) == "" || len(privateKey) != ed25519.PrivateKeySize {
		return "", errors.New("activation signing input is invalid")
	}
	payload, err := json.Marshal(activationCodeEnvelope{Version: ActivationCodeSchemaVersion, KeyID: keyID, Algorithm: "Ed25519", Grant: grant})
	if err != nil {
		return "", err
	}
	code := ActivationCodeSchemaVersion + "." + base64.RawURLEncoding.EncodeToString(payload) + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(privateKey, payload))
	if len(code) > ActivationCodeMaxLength {
		return "", errors.New("activation code exceeds maximum length")
	}
	return code, nil
}

func VerifyActivationCode(code string, keys map[string]ed25519.PublicKey, binding ActivationCodeBinding, now time.Time, nonceSeen func(string) bool) (ActivationGrant, error) {
	parts := strings.Split(code, ".")
	if len(parts) != 3 || parts[0] != ActivationCodeSchemaVersion || len(code) > ActivationCodeMaxLength {
		return ActivationGrant{}, errors.New("invalid activation code")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ActivationGrant{}, errors.New("invalid activation code")
	}
	var envelope activationCodeEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil || envelope.Version != ActivationCodeSchemaVersion || envelope.Algorithm != "Ed25519" {
		return ActivationGrant{}, errors.New("invalid activation code")
	}
	key, ok := keys[envelope.KeyID]
	signature, signatureErr := base64.RawURLEncoding.DecodeString(parts[2])
	if !ok || signatureErr != nil || !ed25519.Verify(key, payload, signature) {
		return ActivationGrant{}, errors.New("activation code signature rejected")
	}
	if err := envelope.Grant.Validate(now, binding); err != nil {
		return ActivationGrant{}, err
	}
	if nonceSeen != nil && nonceSeen(envelope.Grant.Nonce) {
		return ActivationGrant{}, fmt.Errorf("activation code replay rejected")
	}
	return envelope.Grant, nil
}
