package domain

import (
	"fmt"
	"strings"
)

type CleanupPhase string

const (
	CleanupRequested        CleanupPhase = "requested"
	CleanupBackendDeleting  CleanupPhase = "backend_deleting"
	CleanupWaitingFinalizer CleanupPhase = "waiting_finalizers"
	CleanupVerifiedEmpty    CleanupPhase = "verified_empty"
	CleanupTerminated       CleanupPhase = "terminated"
	CleanupFailed           CleanupPhase = "failed"
)

type CleanupState struct {
	Phase      CleanupPhase               `json:"phase"`
	Inventory  []ReleasePlanInventoryItem `json:"inventory,omitempty"`
	Attempts   int                        `json:"attempts,omitempty"`
	Verified   bool                       `json:"verified"`
	Finalizers []string                   `json:"finalizers,omitempty"`
	LastError  string                     `json:"lastError,omitempty"`
}

func (s *CleanupState) Advance(next CleanupPhase) error {
	if s.Phase == next {
		return nil
	}
	allowed := map[CleanupPhase]map[CleanupPhase]bool{
		"":                      {CleanupRequested: true},
		CleanupRequested:        {CleanupBackendDeleting: true, CleanupFailed: true},
		CleanupBackendDeleting:  {CleanupWaitingFinalizer: true, CleanupFailed: true},
		CleanupWaitingFinalizer: {CleanupWaitingFinalizer: true, CleanupVerifiedEmpty: true, CleanupFailed: true},
		CleanupVerifiedEmpty:    {CleanupTerminated: true},
		CleanupFailed:           {CleanupRequested: true, CleanupBackendDeleting: true},
		CleanupTerminated:       {},
	}
	if !allowed[s.Phase][next] {
		return fmt.Errorf("invalid cleanup transition %q -> %q", s.Phase, next)
	}
	s.Phase = next
	return nil
}

func CleanupInventory(plan EnvironmentReleasePlan) []ReleasePlanInventoryItem {
	owned := map[string]bool{}
	for _, record := range plan.Ownership {
		owned[record.Kind+"/"+record.Namespace+"/"+record.Name] = true
	}
	items := make([]ReleasePlanInventoryItem, 0, len(plan.RenderedResources))
	for _, resource := range plan.RenderedResources {
		key := resource.Kind + "/" + resource.Namespace + "/" + resource.Name
		if owned[key] {
			items = append(items, ReleasePlanInventoryItem{Kind: resource.Kind, Namespace: resource.Namespace, Name: resource.Name, Digest: resource.Digest, Owned: true})
		}
	}
	return items
}

func ValidateCleanupInventory(items []ReleasePlanInventoryItem, tenant, project, environment string) error {
	if strings.TrimSpace(tenant) == "" || strings.TrimSpace(project) == "" || strings.TrimSpace(environment) == "" {
		return fmt.Errorf("cleanup identity is incomplete")
	}
	seen := map[string]bool{}
	for _, item := range items {
		key := item.Kind + "/" + item.Namespace + "/" + item.Name
		if seen[key] || !item.Owned || strings.TrimSpace(item.Kind) == "" || strings.TrimSpace(item.Namespace) == "" || strings.TrimSpace(item.Name) == "" {
			return fmt.Errorf("invalid cleanup inventory item %s", key)
		}
		seen[key] = true
		if item.Kind == "Namespace" {
			if item.Name != environment {
				return fmt.Errorf("cleanup cannot delete source namespace")
			}
		} else if item.Namespace != environment {
			return fmt.Errorf("cleanup inventory crosses namespace boundary")
		}
	}
	return nil
}
