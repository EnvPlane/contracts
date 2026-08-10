package domain

import "testing"

func TestTenantPrincipalRequiresTenantAndSubject(t *testing.T) {
	if (TenantPrincipal{TenantID: "t", Subject: "s"}).Valid() == false {
		t.Fatal("expected valid principal")
	}
	if (TenantPrincipal{TenantID: "t"}).Valid() || (TenantPrincipal{Subject: "s"}).Valid() {
		t.Fatal("incomplete principal must be invalid")
	}
}
