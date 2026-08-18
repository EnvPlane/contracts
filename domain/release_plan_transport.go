package domain

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

type ReleasePlanInventoryItem struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Digest    string `json:"digest"`
	Owned     bool   `json:"owned"`
}
type ReleasePlanRunnerIdentity struct {
	TenantID  string `json:"tenantId"`
	ProjectID string `json:"projectId"`
	ClusterID string `json:"clusterId,omitempty"`
	RunnerID  string `json:"runnerId,omitempty"`
}
type ReleasePlanTransportReference struct {
	PlanID         string `json:"planId"`
	PlanDigest     string `json:"planDigest"`
	TemplateDigest string `json:"templateDigest"`
	InputDigest    string `json:"inputDigest"`
	Signature      string `json:"signature"`
	KeyID          string `json:"keyId"`
}

func (p EnvironmentReleasePlan) Inventory() []ReleasePlanInventoryItem {
	items := make([]ReleasePlanInventoryItem, 0, len(p.RenderedResources))
	owned := map[string]bool{}
	for _, record := range p.Ownership {
		owned[record.Kind+"/"+record.Namespace+"/"+record.Name] = true
	}
	for _, resource := range p.RenderedResources {
		items = append(items, ReleasePlanInventoryItem{Kind: resource.Kind, Namespace: resource.Namespace, Name: resource.Name, Digest: resource.Digest, Owned: owned[resource.Kind+"/"+resource.Namespace+"/"+resource.Name]})
	}
	sort.Slice(items, func(i, j int) bool { return inventoryKey(items[i]) < inventoryKey(items[j]) })
	return items
}

func (p EnvironmentReleasePlan) ValidateForExecution(identity ReleasePlanRunnerIdentity, namespaces, kinds []string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if identity.TenantID != p.TenantID || identity.ProjectID != p.ProjectID {
		return errors.New("release plan tenant/project identity mismatch")
	}
	if len(p.RenderedResources) == 0 {
		return errors.New("release plan inventory is empty")
	}
	seen := map[string]bool{}
	for _, resource := range p.RenderedResources {
		key := resource.Kind + "/" + resource.Namespace + "/" + resource.Name
		if seen[key] || !containsExact(namespaces, resource.Namespace) || !containsExact(kinds, resource.Kind) {
			return fmt.Errorf("release plan resource is outside exact inventory guard: %s", key)
		}
		seen[key] = true
		if !manifestIdentityMatches(resource.Manifest, resource.Kind, resource.Namespace, resource.Name) {
			return fmt.Errorf("release plan manifest identity mismatch: %s", key)
		}
		if containsInlineCredential(resource.Manifest) {
			return fmt.Errorf("release plan contains inline credentials: %s", key)
		}
	}
	if len(p.Ownership) != len(p.RenderedResources) {
		return errors.New("release plan ownership inventory is incomplete")
	}
	for _, item := range p.Inventory() {
		if !item.Owned {
			return fmt.Errorf("release plan resource is not owned: %s", inventoryKey(item))
		}
	}
	return nil
}
func (p EnvironmentReleasePlan) ReadyFromInventory(observed []ReleasePlanInventoryItem) error {
	want := p.Inventory()
	sort.Slice(observed, func(i, j int) bool { return inventoryKey(observed[i]) < inventoryKey(observed[j]) })
	if len(want) != len(observed) {
		return fmt.Errorf("ready inventory incomplete: expected %d resources, observed %d", len(want), len(observed))
	}
	for i := range want {
		if want[i] != observed[i] {
			return fmt.Errorf("ready inventory mismatch at %s", inventoryKey(want[i]))
		}
	}
	return nil
}
func EquivalentReleasePlans(left, right EnvironmentReleasePlan) error {
	if left.TenantID != right.TenantID || left.ProjectID != right.ProjectID || left.EnvironmentID != right.EnvironmentID {
		return errors.New("release plan identity differs")
	}
	a, b := left.Inventory(), right.Inventory()
	if len(a) != len(b) {
		return errors.New("release plan inventories differ")
	}
	for i := range a {
		if a[i] != b[i] {
			return fmt.Errorf("release plan inventory differs at %s", inventoryKey(a[i]))
		}
	}
	return nil
}
func SignReleasePlanReference(plan EnvironmentReleasePlan, keyID string, privateKey ed25519.PrivateKey) (ReleasePlanTransportReference, error) {
	if err := plan.Validate(); err != nil {
		return ReleasePlanTransportReference{}, err
	}
	if len(privateKey) != ed25519.PrivateKeySize || strings.TrimSpace(keyID) == "" {
		return ReleasePlanTransportReference{}, errors.New("signing key and key id are required")
	}
	return ReleasePlanTransportReference{PlanID: plan.PlanID, PlanDigest: plan.Digest, TemplateDigest: plan.TemplateDigest, InputDigest: plan.InputDigest, Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(privateKey, releasePlanSigningPayload(plan))), KeyID: keyID}, nil
}
func VerifyReleasePlanReference(plan EnvironmentReleasePlan, ref ReleasePlanTransportReference, publicKey ed25519.PublicKey, identity ReleasePlanRunnerIdentity, namespaces, kinds []string) error {
	if ref.PlanID != plan.PlanID || ref.PlanDigest != plan.Digest || ref.TemplateDigest != plan.TemplateDigest || ref.InputDigest != plan.InputDigest {
		return errors.New("release plan content address mismatch")
	}
	signature, err := base64.StdEncoding.DecodeString(ref.Signature)
	if err != nil || len(publicKey) != ed25519.PublicKeySize || !ed25519.Verify(publicKey, releasePlanSigningPayload(plan), signature) {
		return errors.New("release plan signature verification failed")
	}
	return plan.ValidateForExecution(identity, namespaces, kinds)
}
func releasePlanSigningPayload(plan EnvironmentReleasePlan) []byte {
	copy := plan
	copy.Digest = ""
	payload, _ := json.Marshal(copy)
	return payload
}
func inventoryKey(item ReleasePlanInventoryItem) string {
	return item.Kind + "/" + item.Namespace + "/" + item.Name + "/" + item.Digest
}
func containsExact(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
func manifestIdentityMatches(manifest map[string]any, kind, namespace, name string) bool {
	if fmt.Sprint(manifest["kind"]) != kind {
		return false
	}
	metadata, _ := manifest["metadata"].(map[string]any)
	return fmt.Sprint(metadata["name"]) == name && (kind == "Namespace" || fmt.Sprint(metadata["namespace"]) == namespace)
}
func containsInlineCredential(value any) bool {
	switch item := value.(type) {
	case map[string]any:
		for key, child := range item {
			lower := strings.ToLower(key)
			if (lower == "data" || lower == "stringdata") && strings.EqualFold(fmt.Sprint(item["kind"]), "secret") {
				return true
			}
			if lower == "password" || lower == "clientsecret" || lower == "token" {
				if text, ok := child.(string); ok && strings.TrimSpace(text) != "" {
					return true
				}
			}
			if containsInlineCredential(child) {
				return true
			}
		}
	case []any:
		for _, child := range item {
			if containsInlineCredential(child) {
				return true
			}
		}
	}
	return false
}
