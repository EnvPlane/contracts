package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type ReleaseGateStage string

const (
	ReleaseGateScan       ReleaseGateStage = "scan"
	ReleaseGateTemplate   ReleaseGateStage = "template"
	ReleaseGateSCMOpen    ReleaseGateStage = "scm_open"
	ReleaseGateApply      ReleaseGateStage = "apply"
	ReleaseGateParity     ReleaseGateStage = "parity"
	ReleaseGateUpdate     ReleaseGateStage = "update"
	ReleaseGateCleanup    ReleaseGateStage = "cleanup"
	ReleaseGateTerminated ReleaseGateStage = "terminated"
)

type ReleaseGate struct {
	Stage             ReleaseGateStage           `json:"stage"`
	Provider          Provider                   `json:"provider"`
	Backend           DeploymentBackend          `json:"backend"`
	TenantID          string                     `json:"tenantId"`
	ProjectID         string                     `json:"projectId"`
	EnvironmentID     string                     `json:"environmentId"`
	TemplateDigest    string                     `json:"templateDigest,omitempty"`
	PlanDigest        string                     `json:"planDigest,omitempty"`
	DesiredRevision   EnvironmentRevision        `json:"desiredRevision,omitempty"`
	AppliedRevision   EnvironmentRevision        `json:"appliedRevision,omitempty"`
	ExpectedInventory []ReleasePlanInventoryItem `json:"expectedInventory,omitempty"`
	ObservedInventory FeatureInventoryReport     `json:"observedInventory,omitempty"`
	ExpectedGraph     []ObservedDependencyEdge   `json:"expectedGraph,omitempty"`
	ObservedGraph     []ObservedDependencyEdge   `json:"observedGraph,omitempty"`
	Cleanup           CleanupState               `json:"cleanup,omitempty"`
	SCMOpen           bool                       `json:"scmOpen"`
	BackendApplied    bool                       `json:"backendApplied"`
}

var releaseDigestPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

func (g ReleaseGate) Validate() error {
	if strings.TrimSpace(g.TenantID) == "" || strings.TrimSpace(g.ProjectID) == "" || strings.TrimSpace(g.EnvironmentID) == "" {
		return errors.New("release gate identity is incomplete")
	}
	if g.Provider != ProviderGitHub && g.Provider != ProviderGitLab {
		return fmt.Errorf("unsupported release gate provider %q", g.Provider)
	}
	if g.Backend != DeploymentBackendHelmDirect && g.Backend != DeploymentBackendFluxCD && g.Backend != DeploymentBackendArgoCD {
		return fmt.Errorf("unsupported release gate backend %q", g.Backend)
	}
	if g.Stage == "" {
		return errors.New("release gate stage is required")
	}
	if g.TemplateDigest != "" && !releaseDigestPattern.MatchString(strings.ToLower(g.TemplateDigest)) {
		return errors.New("template digest must be immutable")
	}
	if g.PlanDigest != "" && !releaseDigestPattern.MatchString(strings.ToLower(g.PlanDigest)) {
		return errors.New("release plan digest must be immutable")
	}
	if err := noReleasePlaceholders(g); err != nil {
		return err
	}
	if err := noReleaseCredentials(g); err != nil {
		return err
	}
	if releaseGateRank(g.Stage) >= releaseGateRank(ReleaseGateScan) && !g.ObservedInventory.Complete {
		return errors.New("release gate requires complete multi-namespace inventory")
	}
	if releaseGateRank(g.Stage) >= releaseGateRank(ReleaseGateSCMOpen) && !g.SCMOpen {
		return errors.New("release gate requires SCM MR/PR open")
	}
	if releaseGateRank(g.Stage) >= releaseGateRank(ReleaseGateApply) && !g.BackendApplied {
		return errors.New("release gate requires backend apply")
	}
	if releaseGateRank(g.Stage) >= releaseGateRank(ReleaseGateParity) {
		diff, err := CompareFeatureInventory(g.ExpectedInventory, g.ObservedInventory)
		if err != nil {
			return err
		}
		if !diff.Safe {
			return errors.New("release gate inventory parity failed")
		}
		if !sameDependencyGraph(g.ExpectedGraph, g.ObservedGraph) {
			return errors.New("release gate dependency graph parity failed")
		}
	}
	if g.Stage == ReleaseGateTerminated && g.Cleanup.Phase != CleanupTerminated {
		return errors.New("release gate terminated requires verified cleanup")
	}
	return nil
}

func noReleasePlaceholders(value any) error {
	b, _ := json.Marshal(value)
	text := strings.ToLower(string(b))
	for _, marker := range []string{"{{", "}}", "placeholder", "stale artifact", "fixture-only"} {
		if strings.Contains(text, marker) {
			return fmt.Errorf("release gate contains forbidden marker %q", marker)
		}
	}
	return nil
}
func noReleaseCredentials(value any) error {
	b, _ := json.Marshal(value)
	text := strings.ToLower(string(b))
	for _, marker := range []string{"password", "access_token", "refresh_token", "private_key", "client_secret", "bearer "} {
		if strings.Contains(text, marker) {
			return fmt.Errorf("release gate contains credential marker %q", marker)
		}
	}
	return nil
}
func sameDependencyGraph(left, right []ObservedDependencyEdge) bool {
	a := append([]ObservedDependencyEdge(nil), left...)
	b := append([]ObservedDependencyEdge(nil), right...)
	sort.Slice(a, func(i, j int) bool { return edgeKey(a[i]) < edgeKey(a[j]) })
	sort.Slice(b, func(i, j int) bool { return edgeKey(b[i]) < edgeKey(b[j]) })
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
func edgeKey(edge ObservedDependencyEdge) string {
	return edge.From + "\x00" + edge.To + "\x00" + edge.Type + "\x00" + fmt.Sprint(edge.Required)
}

func releaseGateRank(stage ReleaseGateStage) int {
	for rank, candidate := range []ReleaseGateStage{ReleaseGateScan, ReleaseGateTemplate, ReleaseGateSCMOpen, ReleaseGateApply, ReleaseGateParity, ReleaseGateUpdate, ReleaseGateCleanup, ReleaseGateTerminated} {
		if stage == candidate { return rank }
	}
	return -1
}
