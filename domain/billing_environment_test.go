package domain

import "testing"

func TestVerifiedWebhookRequiresConfiguredEnvironment(t *testing.T) {
	for _, mode := range []string{"test", "live"} {
		raw := []byte(`{"id":"evt_owned","type":"invoice.paid","livemode":` + map[string]string{"test": "false", "live": "true"}[mode] + `,"data":{"object":{"id":"in_owned"}}}`)
		e, err := ParseVerifiedBillingWebhook(raw, mode)
		if err != nil || e.Mode != mode || e.ObjectID != "in_owned" {
			t.Fatal("configured webhook rejected")
		}
		for _, wrong := range []string{"", "unknown", map[string]string{"test": "live", "live": "test"}[mode]} {
			if _, err := ParseVerifiedBillingWebhook(raw, wrong); err == nil {
				t.Fatal("wrong environment accepted")
			}
		}
		owner := BillingOwner{BillingSellerNamespace, mode, "tenant", "installation"}
		if owner.ValidateMode(mode) != nil {
			t.Fatal("environment owner rejected")
		}
		if mode == "live" && owner.Validate() == nil {
			t.Fatal("legacy owner admitted live")
		}
	}
	if _, err := ParseVerifiedTestBillingWebhook([]byte(`{"id":"evt_live","type":"invoice.paid","livemode":true,"data":{"object":{"id":"in_live"}}}`)); err == nil {
		t.Fatal("legacy test parser accepted live")
	}
	for _, raw := range []string{`{"id":"evt_owned","type":"invoice.paid","livemode":false,"livemode":true,"data":{"object":{"id":"in_owned"}}}`, `{"id":"evt_owned","type":"invoice.paid","LiveMode":true,"data":{"object":{"id":"in_owned"}}}`} {
		if _, err := ParseVerifiedBillingWebhook([]byte(raw), "live"); err == nil {
			t.Fatal("ambiguous live envelope accepted")
		}
	}
}
