package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestBillingWebhookRawSignatureReplayAndEnvelope(t *testing.T) {
	now := time.Unix(2000000000, 0)
	raw := []byte(`{"id":"evt_fixture","type":"invoice.paid","livemode":false,"data":{"object":{"id":"in_fixture","metadata":{"tenantId":"foreign","email":"private@example.invalid"}}}}`)
	sign := func(raw []byte, at time.Time, key string) string {
		stamp := strconv.FormatInt(at.Unix(), 10)
		mac := hmac.New(sha256.New, []byte(key))
		mac.Write([]byte(stamp + "."))
		mac.Write(raw)
		return "t=" + stamp + ",v1=" + hex.EncodeToString(mac.Sum(nil))
	}
	header := sign(raw, now, "fixture")
	if err := VerifyBillingWebhookSignature(raw, header, "fixture", now, 5*time.Minute); err != nil {
		t.Fatal(err)
	}
	e, err := ParseVerifiedTestBillingWebhook(raw)
	if err != nil || e.Mode != "test" || e.EventID != "evt_fixture" || strings.Contains(e.PayloadHash, "private") {
		t.Fatal(e, err)
	}
	for _, tc := range []struct {
		body        []byte
		header, key string
	}{{append(raw, ' '), header, "fixture"}, {raw, header, "live-key"}, {raw, sign(raw, now.Add(-6*time.Minute), "fixture"), "fixture"}, {raw, sign(raw, now.Add(6*time.Minute), "fixture"), "fixture"}, {raw, header + ",t=2000000000", "fixture"}, {raw, "t=," + header, "fixture"}, {[]byte(strings.Repeat("x", BillingWebhookMaxBytes+1)), header, "fixture"}} {
		if err := VerifyBillingWebhookSignature(tc.body, tc.header, tc.key, now, 5*time.Minute); err == nil {
			t.Fatal("invalid signature accepted")
		}
	}
	for _, body := range []string{`{"id":"evt_x","type":"invoice.paid","data":{"object":{"id":"in_x"}}}`, `{"id":"evt_x","type":"invoice.paid","livemode":true,"data":{"object":{"id":"in_x"}}}`, `{"id":"evt_x","type":"invoice.paid","livemode":true,"livemode":false,"data":{"object":{"id":"in_x"}}}`, `{"id":"evt_x","type":"invoice.paid","livemode":true,"LiveMode":false,"data":{"object":{"id":"in_x"}}}`, `{"id":"evt_x","type":"invoice.paid","livemode":false,"data":{"object":{"id":"in_x/foreign"}}}`} {
		if _, err := ParseVerifiedTestBillingWebhook([]byte(body)); err == nil {
			t.Fatal("invalid envelope accepted")
		}
	}
	ignored, err := ParseVerifiedTestBillingWebhook([]byte(`{"id":"evt_ignored","type":"customer.created","livemode":false,"data":{"object":{"id":"private@example.invalid"}}}`))
	if err != nil || ignored.ObjectID != "" {
		t.Fatal("ignored event stores PII", ignored, err)
	}
}
