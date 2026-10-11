package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"
)

const BillingWebhookMaxBytes = 1 << 20
const BillingWebhookProvider = "stripe"
const BillingWebhookMode = "test"

// Verify raw bytes before any JSON parsing; multiple v1 values allow key rotation.
func VerifyBillingWebhookSignature(raw []byte, header, secret string, now time.Time, tolerance time.Duration) error {
	reject := errors.New("billing webhook signature rejected")
	if secret == "" || len(raw) > BillingWebhookMaxBytes || len(header) > 4096 || tolerance <= 0 || tolerance > 5*time.Minute {
		return reject
	}
	var stamp string
	var haveStamp bool
	var signatures []string
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "t" {
			if haveStamp {
				return reject
			}
			haveStamp = true
			stamp = kv[1]
		}
		if kv[0] == "v1" {
			signatures = append(signatures, kv[1])
		}
	}
	seconds, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil || len(signatures) > 16 {
		return reject
	}
	at := time.Unix(seconds, 0)
	if at.Before(now.Add(-tolerance)) || at.After(now.Add(tolerance)) {
		return reject
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stamp + "."))
	_, _ = mac.Write(raw)
	for _, signature := range signatures {
		got, err := hex.DecodeString(signature)
		if err == nil && hmac.Equal(got, mac.Sum(nil)) {
			return nil
		}
	}
	return reject
}
func SupportedBillingWebhook(kind string) bool {
	switch kind {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded", "checkout.session.async_payment_failed", "checkout.session.expired", "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted", "invoice.paid", "invoice.payment_succeeded", "invoice.payment_failed", "invoice.updated":
		return true
	}
	return false
}

// The envelope stores identifiers and a raw-body digest only, never payload/PII.
type VerifiedBillingWebhook struct{ Provider, Mode, EventID, Kind, ObjectID, PayloadHash string }

func ParseVerifiedTestBillingWebhook(raw []byte) (VerifiedBillingWebhook, error) {
	reject := errors.New("billing webhook envelope rejected")
	if len(raw) > BillingWebhookMaxBytes || !uniqueBillingJSON(raw) {
		return VerifiedBillingWebhook{}, reject
	}
	var e struct {
		ID   string `json:"id"`
		Kind string `json:"type"`
		Live *bool  `json:"livemode"`
		Data struct {
			Object struct {
				ID string `json:"id"`
			} `json:"object"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &e) != nil || e.Live == nil || *e.Live || !billingEventIdentifier(e.ID, 128) || !billingEventIdentifier(e.Kind, 128) || (SupportedBillingWebhook(e.Kind) && !billingEventIdentifier(e.Data.Object.ID, 128)) {
		return VerifiedBillingWebhook{}, reject
	}
	if !SupportedBillingWebhook(e.Kind) {
		e.Data.Object.ID = ""
	}
	hash := sha256.Sum256(raw)
	return VerifiedBillingWebhook{BillingWebhookProvider, BillingWebhookMode, e.ID, e.Kind, e.Data.Object.ID, hex.EncodeToString(hash[:])}, nil
}
func billingEventIdentifier(s string, max int) bool {
	if s == "" || len(s) > max {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '.' || r == '-') {
			return false
		}
	}
	return true
}

// Signed but ambiguous JSON must not bypass the explicit environment boundary.
func uniqueBillingJSON(raw []byte) bool {
	d := json.NewDecoder(strings.NewReader(string(raw)))
	var value func(int, string) error
	value = func(depth int, path string) error {
		if depth > 64 {
			return errors.New("depth")
		}
		token, err := d.Token()
		if err != nil {
			return err
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			keys := map[string]bool{}
			for d.More() {
				k, err := d.Token()
				key, ok := k.(string)
				if err != nil || !ok || keys[key] {
					return errors.New("duplicate")
				}
				lower := strings.ToLower(key)
				critical := (path == "" && (lower == "id" || lower == "type" || lower == "livemode" || lower == "data")) || (path == "/data" && lower == "object") || (path == "/data/object" && lower == "id")
				if critical && key != lower {
					return errors.New("ambiguous envelope")
				}
				keys[key] = true
				if err := value(depth+1, path+"/"+key); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim('}') {
				return errors.New("object")
			}
		case '[':
			for d.More() {
				if err := value(depth+1, path+"/*"); err != nil {
					return err
				}
			}
			end, err := d.Token()
			if err != nil || end != json.Delim(']') {
				return errors.New("array")
			}
		default:
			return errors.New("delimiter")
		}
		return nil
	}
	if value(0, "") != nil {
		return false
	}
	return d.Decode(new(any)) == io.EOF
}
