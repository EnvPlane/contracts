package domain

import (
	"strings"
	"testing"
	"time"
)

func TestBillingRequestScope(t *testing.T) {
	now := time.Unix(2000000000, 0)
	owner := BillingOwner{BillingSellerNamespace, "test", "tenant-a", "installation-a"}
	body := []byte(`{"planId":"team"}`)
	header, err := SignBillingRequestScope(owner, "POST", "/v1/internal/billing/checkout", body, "fixture", now)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(header, ".8zBKgUzJTZ-UEmghYa5Z5R0TCFgEp9dhHdRw0GP2RBo") {
		t.Fatal("operator protocol drift")
	}

	got, err := VerifyBillingRequestScope(header, "fixture", "POST", "/v1/internal/billing/checkout", body, now)
	if err != nil || got != owner {
		t.Fatal(got, err)
	}
	for _, tc := range []struct {
		name, header, secret, method, target string
		body                                 []byte
		at                                   time.Time
	}{
		{"missing", "", "fixture", "POST", "/v1/internal/billing/checkout", body, now},
		{"signature", header, "other", "POST", "/v1/internal/billing/checkout", body, now},
		{"body", header, "fixture", "POST", "/v1/internal/billing/checkout", []byte(`{"tenantId":"foreign"}`), now},
		{"route", header, "fixture", "POST", "/v1/internal/billing/portal", body, now},
		{"method", header, "fixture", "GET", "/v1/internal/billing/checkout", body, now},
		{"expired", header, "fixture", "POST", "/v1/internal/billing/checkout", body, now.Add(2 * time.Minute)},
		{"future", header, "fixture", "POST", "/v1/internal/billing/checkout", body, now.Add(-time.Second)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := VerifyBillingRequestScope(tc.header, tc.secret, tc.method, tc.target, tc.body, tc.at); err == nil {
				t.Fatal("invalid scope accepted")
			}
		})
	}
	for _, bad := range []BillingOwner{{BillingSellerNamespace, "live", "t", "i"}, {"other", "test", "t", "i"}, {BillingSellerNamespace, "test", "", "i"}, {BillingSellerNamespace, "test", "t\n", "i"}} {
		if _, err := SignBillingRequestScope(bad, "POST", "/v1/internal/billing/customer", nil, "fixture", now); err == nil {
			t.Fatal("invalid owner accepted")
		}
	}
}
