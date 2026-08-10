package domain

import (
	"fmt"
	"sort"
	"strings"
)

const PlanSchemaVersion = "1"

type PlanDefinition struct {
	ID               string           `json:"id"`
	SchemaVersion    string           `json:"schemaVersion"`
	EffectiveVersion string           `json:"effectiveVersion"`
	Features         map[string]bool  `json:"features"`
	Limits           map[string]int64 `json:"limits"`
}

type PlanCatalog struct {
	SchemaVersion string           `json:"schemaVersion"`
	Plans         []PlanDefinition `json:"plans"`
}

func CommunityFreePlanCatalog() PlanCatalog {
	return PlanCatalog{
		SchemaVersion: PlanSchemaVersion,
		Plans: []PlanDefinition{
			{ID: "community", SchemaVersion: PlanSchemaVersion, EffectiveVersion: "1.0.0", Features: map[string]bool{
				"projects": true, "environments": true, "gitops": true, "helmDirect": true, "audit": true,
			}, Limits: map[string]int64{"maxProjects": 10, "maxActiveEnvironments": 25, "maxRemoteClusters": 3, "maxMembers": 10, "maxTTLHours": 720, "maxPinDays": 30}},
			{ID: "free", SchemaVersion: PlanSchemaVersion, EffectiveVersion: "1.0.0", Features: map[string]bool{
				"projects": true, "environments": true, "gitops": true, "helmDirect": false, "audit": true,
			}, Limits: map[string]int64{"maxProjects": 3, "maxActiveEnvironments": 2, "maxRemoteClusters": 1, "maxMembers": 3, "maxTTLHours": 72, "maxPinDays": 7}},
		},
	}
}

func (c PlanCatalog) Validate() error {
	if strings.TrimSpace(c.SchemaVersion) == "" {
		return fmt.Errorf("plan catalog schema version is required")
	}
	seen := map[string]struct{}{}
	for _, plan := range c.Plans {
		if strings.TrimSpace(plan.ID) == "" || strings.TrimSpace(plan.EffectiveVersion) == "" {
			return fmt.Errorf("plan id and effective version are required")
		}
		if plan.SchemaVersion != c.SchemaVersion {
			return fmt.Errorf("plan %q schema version mismatch", plan.ID)
		}
		if _, ok := seen[plan.ID]; ok {
			return fmt.Errorf("duplicate plan %q", plan.ID)
		}
		seen[plan.ID] = struct{}{}
		for key, limit := range plan.Limits {
			if strings.TrimSpace(key) == "" || limit < 0 {
				return fmt.Errorf("invalid limit %q in plan %q", key, plan.ID)
			}
		}
		for key := range plan.Features {
			if strings.TrimSpace(key) == "" {
				return fmt.Errorf("empty feature key in plan %q", plan.ID)
			}
		}
	}
	return nil
}

func (c PlanCatalog) Deterministic() PlanCatalog {
	copyCatalog := c
	sort.Slice(copyCatalog.Plans, func(i, j int) bool { return copyCatalog.Plans[i].ID < copyCatalog.Plans[j].ID })
	return copyCatalog
}
