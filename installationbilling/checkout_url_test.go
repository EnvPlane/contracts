package installationbilling

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCheckoutOpaqueURLFragments(t *testing.T) {
	var vectors struct {
		Cases []struct {
			Name, Operation, URL string
			Valid                bool
		}
	}
	raw, err := os.ReadFile("testdata/checkout-url-v1.json")
	if err != nil || json.Unmarshal(raw, &vectors) != nil || len(vectors.Cases) != 25 {
		t.Fatal("synthetic URL vectors unavailable")
	}
	corpus := loadCorpus(t)
	for _, vector := range vectors.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			for _, p := range corpus.Positive {
				request, err := ParseVerifiedRequest([]byte(p.Request), 1800000000)
				if err != nil {
					t.Fatal("existing proof rejected")
				}
				if request.Intent.Operation != vector.Operation {
					continue
				}
				var response Response
				if json.Unmarshal([]byte(p.Response), &response) != nil {
					t.Fatal("existing response unavailable")
				}
				response.URL = vector.URL
				encoded, _ := json.Marshal(response)
				parsed, err := ParseResponse(encoded)
				if vector.Valid {
					if err != nil || parsed.URL != vector.URL || request.MatchResponse(parsed) != nil {
						t.Fatal("opaque URL or request binding changed")
					}
					parsed.RequestID = b64(make([]byte, 16))
					if request.MatchResponse(parsed) == nil {
						t.Fatal("foreign response accepted")
					}
				} else if err == nil {
					t.Fatal("unsafe synthetic hosted URL accepted")
				}
				return
			}
			t.Fatal("matching proof fixture unavailable")
		})
	}
	for _, count := range []int{4096, 4097} {
		p := corpus.Positive[0]
		for _, candidate := range corpus.Positive {
			if candidate.Name == "checkout" {
				p = candidate
				break
			}
		}
		var response Response
		_ = json.Unmarshal([]byte(p.Response), &response)
		prefix := "https://checkout.stripe.com/c/pay/offline-fixture#"
		response.URL = prefix + strings.Repeat("x", count-len(prefix))
		encoded, _ := json.Marshal(response)
		_, err := ParseResponse(encoded)
		if (err == nil) != (count == 4096) {
			t.Fatal("hosted URL length boundary changed")
		}
	}
	for _, operation := range []string{"checkout", "portal"} {
		body := map[string]any{"returnRoute": "billing", "returnURL": "https://seller.invalid/billing#opaque-fixture"}
		if operation == "checkout" {
			body["attemptId"], body["planId"], body["planVersion"] = "fixture-attempt", "team", "fixture-version"
		}
		encoded, _ := json.Marshal(body)
		if ValidateBody(operation, encoded) == nil {
			t.Fatal("caller return URL accepted")
		}
		delete(body, "returnURL")
		body["returnRoute"] = "billing#opaque-fixture"
		encoded, _ = json.Marshal(body)
		if ValidateBody(operation, encoded) == nil {
			t.Fatal("return route fragment accepted")
		}
	}
}
