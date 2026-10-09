package domain

import (
	"encoding/json"
	"testing"
)

func TestScaledRateBoundsAndLegacySafety(t *testing.T) {
	for _, rate := range []FinOpsScaledRate{{0, 0}, {5, 1}, {14, 3}, {1_000_000_000_000_000, 6}} {
		if !rate.Valid() {
			t.Fatalf("valid rate rejected: %+v", rate)
		}
	}
	for _, rate := range []FinOpsScaledRate{{-1, 0}, {1, -1}, {1, 7}, {1_000_000_001, 0}, {1_000_000_000_000_001, 6}} {
		if rate.Valid() {
			t.Fatalf("unsafe rate accepted: %+v", rate)
		}
	}
	price := FinOpsDimensionPrice{ScaledRate: &FinOpsScaledRate{14, 3}}
	data, err := json.Marshal(price)
	if err != nil {
		t.Fatal(err)
	}
	var old struct {
		MinorUnitsPerUnit int64 `json:"minorUnitsPerUnit"`
		Known             bool  `json:"known"`
	}
	if err = json.Unmarshal(data, &old); err != nil || old.Known {
		t.Fatal("old reader fabricated a known whole-cent/zero price", err)
	}
	for _, invalid := range []string{`{}`, `{"scale":3}`, `{"minorUnits":0}`, `{"minorUnits":null,"scale":0}`, `{"minorUnits":1,"scale":7}`, `{"minorUnits":1,"scale":0,"extra":1}`, `{"minorUnits":1.5,"scale":0}`} {
		var rate FinOpsScaledRate
		if json.Unmarshal([]byte(invalid), &rate) == nil {
			t.Fatalf("missing/malformed rate became known zero: %s", invalid)
		}
	}
}
