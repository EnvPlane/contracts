package installationbilling

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestOpenAPIExactSellerOperationInventory(t *testing.T) {
	raw, e := os.ReadFile("../openapi/openapi.json")
	if e != nil {
		t.Fatal(e)
	}
	var d struct {
		Paths      map[string]map[string]json.RawMessage `json:"paths"`
		Components struct {
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	if json.Unmarshal(raw, &d) != nil {
		t.Fatal("spec")
	}
	wanted := map[string]bool{ChallengePath: true}
	for _, p := range Routes() {
		wanted[p] = true
	}
	seen := 0
	for path, item := range d.Paths {
		if !strings.HasPrefix(path, "/v1/installation-billing/") {
			continue
		}
		if !wanted[path] || len(item) != 1 {
			t.Fatal("unreviewed seller route", path)
		}
		var op struct {
			Host           string `json:"x-envplane-host"`
			Implementation string `json:"x-envplane-implementation"`
			Security       []any  `json:"security"`
			RequestBody    struct {
				Required bool `json:"required"`
			} `json:"requestBody"`
		}
		if json.Unmarshal(item["post"], &op) != nil || op.Host != "seller-issuer" || op.Implementation != "contract-only" || op.Security == nil || len(op.Security) != 0 || !op.RequestBody.Required {
			t.Fatal("scope marker", path)
		}
		seen++
	}
	if seen != 12 {
		t.Fatal("incomplete routes", seen)
	}
	for _, name := range []string{"Scope", "Intent", "Challenge", "Request", "ChallengeRequest", "ChallengeResponse", "Session", "Response", "Record", "Error", "CheckoutBody", "PortalBody", "StatusBody", "KeyStatusBody", "EmptyBody"} {
		var schema struct {
			Additional bool           `json:"additionalProperties"`
			Required   []string       `json:"required"`
			Properties map[string]any `json:"properties"`
		}
		raw := d.Components.Schemas["InstallationBilling"+name]
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &schema) != nil || json.Unmarshal(raw, &fields) != nil || string(fields["additionalProperties"]) != "false" || len(schema.Required) != len(schema.Properties) {
			t.Fatal("open/missing schema", name)
		}
	}
}
