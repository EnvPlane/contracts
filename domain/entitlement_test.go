package domain

import "testing"

func TestEntitlementOverrideValidate(t *testing.T) {
	tests := []struct {
		name     string
		override EntitlementOverride
		wantErr  bool
	}{
		{name: "valid", override: EntitlementOverride{Features: map[string]bool{"scim": true}, Limits: map[string]int64{"projects": 2}}},
		{name: "negative limit", override: EntitlementOverride{Limits: map[string]int64{"projects": -1}}, wantErr: true},
		{name: "empty limit key", override: EntitlementOverride{Limits: map[string]int64{" ": 1}}, wantErr: true},
		{name: "empty feature key", override: EntitlementOverride{Features: map[string]bool{"": true}}, wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.override.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr=%v", err, test.wantErr)
			}
		})
	}
}
