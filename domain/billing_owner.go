package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Each current issuer database belongs to one seller and exactly one provider environment.
// This is an opaque namespace, not a claim about a legal merchant/provider account.
const BillingSellerNamespace = "issuer-default"
const BillingScopeHeader = "X-EnvPlane-Billing-Scope"
const BillingScopeAudience = "envplane-issuer-billing-v1"

type BillingOwner struct {
	SellerID       string `json:"sellerId"`
	ProviderMode   string `json:"providerMode"`
	TenantID       string `json:"tenantId"`
	InstallationID string `json:"installationId"`
}

func (o BillingOwner) Validate() error {
	if o.SellerID != BillingSellerNamespace || (o.ProviderMode != "test" && o.ProviderMode != "live") {
		return errors.New("billing ownership scope rejected")
	}
	for _, s := range []string{o.TenantID, o.InstallationID} {
		if s == "" || len(s) > 128 || strings.TrimSpace(s) != s || strings.ContainsAny(s, "\x00\r\n") {
			return errors.New("billing ownership scope rejected")
		}
	}
	return nil
}

type BillingRequestScope struct {
	Owner     BillingOwner `json:"owner"`
	Audience  string       `json:"audience"`
	Method    string       `json:"method"`
	Target    string       `json:"target"`
	BodyHash  string       `json:"bodyHash"`
	IssuedAt  int64        `json:"issuedAt"`
	ExpiresAt int64        `json:"expiresAt"`
}

func billingBodyHash(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// SignBillingRequestScope is for trusted seller services only. Caller tenant must
// come from authenticated authorization, not arbitrary browser/request metadata.
func SignBillingRequestScope(owner BillingOwner, method, target string, body []byte, secret string, now time.Time) (string, error) {
	if owner.Validate() != nil || secret == "" || (method != "POST" && method != "GET") || !strings.HasPrefix(target, "/v1/internal/billing/") {
		return "", errors.New("billing request scope rejected")
	}
	raw, _ := json.Marshal(BillingRequestScope{owner, BillingScopeAudience, method, target, billingBodyHash(body), now.Unix(), now.Add(2 * time.Minute).Unix()})
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
func VerifyBillingRequestScope(header, secret, method, target string, body []byte, now time.Time) (BillingOwner, error) {
	reject := func() (BillingOwner, error) { return BillingOwner{}, errors.New("billing request scope rejected") }
	if secret == "" || len(header) > 4096 {
		return reject()
	}
	parts := strings.Split(header, ".")
	if len(parts) != 2 {
		return reject()
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return reject()
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return reject()
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return reject()
	}
	var scope BillingRequestScope
	if json.Unmarshal(raw, &scope) != nil || scope.Owner.Validate() != nil || scope.Audience != BillingScopeAudience || scope.Method != method || scope.Target != target || scope.BodyHash != billingBodyHash(body) || scope.IssuedAt > now.Unix() || scope.ExpiresAt <= now.Unix() || scope.ExpiresAt-scope.IssuedAt != 120 || now.Unix()-scope.IssuedAt > 120 {
		return reject()
	}
	return scope.Owner, nil
}
