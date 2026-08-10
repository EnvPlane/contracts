package domain

import (
	"fmt"
	"strings"
	"time"
)

// EntitlementOverride is a bounded, time-limited overlay on plan defaults.
// Licensed overlays have lower priority than tenant-specific overlays.
type EntitlementOverride struct {
	Features  map[string]bool  `json:"features,omitempty"`
	Limits    map[string]int64 `json:"limits,omitempty"`
	Revision  string           `json:"revision,omitempty"`
	Source    string           `json:"source,omitempty"`
	ExpiresAt *time.Time       `json:"expiresAt,omitempty"`
}

func (o EntitlementOverride) Validate() error {
	for key, value := range o.Limits {
		if strings.TrimSpace(key) == "" || value < 0 {
			return fmt.Errorf("invalid entitlement limit %q", key)
		}
	}
	for key := range o.Features {
		if strings.TrimSpace(key) == "" {
			return fmt.Errorf("empty entitlement feature key")
		}
	}
	return nil
}

// EntitlementSnapshot is the immutable-at-boundary result returned by a resolver.
// Maps are defensively copied by resolver implementations; callers must treat them as read-only.
type EntitlementSnapshot struct {
	TenantID    string           `json:"tenantId"`
	PlanID      string           `json:"planId"`
	PlanVersion string           `json:"planVersion"`
	Features    map[string]bool  `json:"features"`
	Limits      map[string]int64 `json:"limits"`
	Revision    string           `json:"revision"`
	ExpiresAt   time.Time        `json:"expiresAt"`
}

func (s EntitlementSnapshot) FeatureEnabled(name string) bool { return s.Features[name] }

func (s EntitlementSnapshot) Limit(name string) (int64, bool) {
	v, ok := s.Limits[name]
	return v, ok
}
