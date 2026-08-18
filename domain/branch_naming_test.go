package domain

import "testing"

func TestBranchEnvironmentNameIsBoundedAndCollisionSafe(t *testing.T) {
	left := BranchEnvironmentNameFor("checkout", "github.com/acme/one", "feature/CMS_Login", "github", "12")
	right := BranchEnvironmentNameFor("checkout", "github.com/acme/two", "feature/CMS_Login", "github", "12")
	if left.DisplayName != "feature-cms-login" || !left.Valid() || len(left.ID) > 63 {
		t.Fatalf("unexpected branch naming: %+v", left)
	}
	if left.ID == right.ID {
		t.Fatal("different repositories must not collide")
	}
}

func TestBranchEnvironmentNameHasLegacyAlias(t *testing.T) {
	name := BranchEnvironmentNameFor("orders", "group/orders", "fix/payment-timeout", "gitlab", "122")
	if len(name.Compatibility) != 1 || name.Compatibility[0] != "mr-122" {
		t.Fatalf("unexpected aliases: %#v", name.Compatibility)
	}
	if BranchEnvironmentNameFor("orders", "group/orders", "feature/"+string(make([]byte, 120)), "gitlab", "122").Valid() == false {
		t.Fatal("bounded names must remain valid")
	}
}
