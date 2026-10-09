package domain

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// FinOpsScaledRate represents minorUnits / 10^scale per resource unit, not a
// rounded Money amount. Integer fields remain exact in JSON and JavaScript.
type FinOpsScaledRate struct {
	MinorUnits int64 `json:"minorUnits"`
	Scale      int32 `json:"scale"`
}

func (r *FinOpsScaledRate) UnmarshalJSON(data []byte) error {
	var wire struct {
		MinorUnits *int64 `json:"minorUnits"`
		Scale      *int32 `json:"scale"`
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(&wire) != nil || !errors.Is(d.Decode(new(any)), io.EOF) || wire.MinorUnits == nil || wire.Scale == nil {
		return errors.New("scaled rate requires explicit integer minorUnits and scale")
	}
	next := FinOpsScaledRate{MinorUnits: *wire.MinorUnits, Scale: *wire.Scale}
	if !next.Valid() {
		return errors.New("scaled rate outside exact supported bounds")
	}
	*r = next
	return nil
}

const MaxFinOpsRateScale int32 = 6

func (r FinOpsScaledRate) Denominator() int64 {
	if r.Scale < 0 || r.Scale > MaxFinOpsRateScale {
		return 0
	}
	d := int64(1)
	for range r.Scale {
		d *= 10
	}
	return d
}

func (r FinOpsScaledRate) Valid() bool {
	d := r.Denominator()
	return d != 0 && r.MinorUnits >= 0 && r.MinorUnits <= 1_000_000_000*d
}
