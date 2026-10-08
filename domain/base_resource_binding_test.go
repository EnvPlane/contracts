package domain

import "testing"

func TestFinOpsResourceAttributionXOR(t *testing.T) {
	for _, tc := range []struct {
		a     FinOpsResourceAttribution
		valid bool
	}{
		{FinOpsResourceAttribution{EnvironmentID: "e"}, true},
		{FinOpsResourceAttribution{BaseResourceBindingID: "b", BindingVersion: 1}, true},
		{FinOpsResourceAttribution{}, false},
		{FinOpsResourceAttribution{EnvironmentID: "e", BaseResourceBindingID: "b", BindingVersion: 1}, false},
		{FinOpsResourceAttribution{BaseResourceBindingID: "b"}, false},
		{FinOpsResourceAttribution{EnvironmentID: "e", BindingVersion: 1}, false},
	} {
		if (tc.a.Validate() == nil) != tc.valid {
			t.Fatalf("XOR %+v", tc)
		}
	}
}
