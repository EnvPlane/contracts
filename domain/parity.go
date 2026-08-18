package domain

import (
	"errors"
	"sort"
	"strings"
)

// ObservedInventoryItem is metadata-only and never carries manifests or data.
type ObservedInventoryItem struct {
	ReleasePlanInventoryItem
	SourceNamespace  string          `json:"sourceNamespace,omitempty"`
	SourceKind       string          `json:"sourceKind,omitempty"`
	SourceName       string          `json:"sourceName,omitempty"`
	Health           *ResourceHealth `json:"health,omitempty"`
	ConfiguredProbes []string        `json:"configuredProbes,omitempty"`
}

type ObservedDependencyEdge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Type     string `json:"type"`
	Required bool   `json:"required,omitempty"`
}

type FeatureInventoryReport struct {
	Complete               bool                     `json:"complete"`
	Items                  []ObservedInventoryItem  `json:"items"`
	Dependencies           []ObservedDependencyEdge `json:"dependencies,omitempty"`
	IndependenceIssues     []DependencyGraphIssue   `json:"independenceIssues,omitempty"`
	SourceHealthDiagnostic []ObservedInventoryItem  `json:"sourceHealthDiagnostic,omitempty"`
	ParityDigest           string                   `json:"parityDigest,omitempty"`
}

type InventoryMismatch struct {
	Expected ReleasePlanInventoryItem `json:"expected"`
	Observed ReleasePlanInventoryItem `json:"observed"`
}
type SafeInventoryDiff struct {
	Expected []ReleasePlanInventoryItem `json:"expected,omitempty"`
	Missing  []ReleasePlanInventoryItem `json:"missing,omitempty"`
	Extra    []ReleasePlanInventoryItem `json:"extra,omitempty"`
	Mismatch []InventoryMismatch        `json:"mismatch,omitempty"`
	Dangling []DependencyGraphIssue     `json:"dangling,omitempty"`
	Safe     bool                       `json:"safe"`
}

type FeatureReadiness struct {
	Ready                bool     `json:"ready"`
	BackendApplyComplete bool     `json:"backendApplyComplete"`
	StructuralParity     bool     `json:"structuralParity"`
	NoForbiddenBaseLinks bool     `json:"noForbiddenBaseLinks"`
	WorkloadsReady       bool     `json:"workloadsReady"`
	ConfiguredProbes     bool     `json:"configuredProbes"`
	StatefulComplete     bool     `json:"statefulComplete"`
	Reasons              []string `json:"reasons,omitempty"`
	ParityDigest         string   `json:"parityDigest,omitempty"`
}

func ObservedInventoryFromSnapshots(snapshots []ResourceSnapshot, graph ServiceGraph) FeatureInventoryReport {
	items := make([]ObservedInventoryItem, 0, len(snapshots))
	for _, snapshot := range snapshots {
		if strings.EqualFold(snapshot.Kind, "Secret") || strings.EqualFold(snapshot.Kind, "Pod") {
			continue
		}
		item := ObservedInventoryItem{ReleasePlanInventoryItem: ReleasePlanInventoryItem{Kind: snapshot.Kind, Namespace: snapshot.Namespace, Name: snapshot.Name, Owned: true}, Health: snapshot.Health}
		if snapshot.SourceMapping != nil {
			item.SourceNamespace, item.SourceKind, item.SourceName = snapshot.SourceMapping.Namespace, snapshot.SourceMapping.Kind, snapshot.SourceMapping.Name
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return inventoryIdentity(items[i].ReleasePlanInventoryItem) < inventoryIdentity(items[j].ReleasePlanInventoryItem)
	})
	edges := make([]ObservedDependencyEdge, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		edges = append(edges, ObservedDependencyEdge{From: edge.From, To: edge.To, Type: edge.Type, Required: edge.Required})
	}
	return FeatureInventoryReport{Complete: true, Items: items, Dependencies: edges}
}

func CompareFeatureInventory(expected []ReleasePlanInventoryItem, observed FeatureInventoryReport) (SafeInventoryDiff, error) {
	if !observed.Complete {
		return SafeInventoryDiff{Expected: append([]ReleasePlanInventoryItem(nil), expected...)}, errors.New("observed feature inventory is incomplete")
	}
	want := append([]ReleasePlanInventoryItem(nil), expected...)
	got := make([]ReleasePlanInventoryItem, 0, len(observed.Items))
	for _, item := range observed.Items {
		got = append(got, item.ReleasePlanInventoryItem)
	}
	sortInventoryItems(want)
	sortInventoryItems(got)
	diff := SafeInventoryDiff{Expected: want, Safe: true}
	wb, gb := inventoryByIdentity(want), inventoryByIdentity(got)
	for key, item := range wb {
		other, ok := gb[key]
		if !ok {
			diff.Missing = append(diff.Missing, item)
		} else if item.Digest != "" && item.Digest != other.Digest {
			diff.Mismatch = append(diff.Mismatch, InventoryMismatch{Expected: item, Observed: other})
		}
	}
	for key, item := range gb {
		if _, ok := wb[key]; !ok {
			diff.Extra = append(diff.Extra, item)
		}
	}
	sortInventoryItems(diff.Missing)
	sortInventoryItems(diff.Extra)
	diff.Safe = len(diff.Missing) == 0 && len(diff.Mismatch) == 0
	return diff, nil
}

func ValidateFeatureReadiness(applyComplete bool, diff SafeInventoryDiff, observed FeatureInventoryReport, statefulComplete bool) FeatureReadiness {
	r := FeatureReadiness{BackendApplyComplete: applyComplete, StructuralParity: diff.Safe, NoForbiddenBaseLinks: len(observed.IndependenceIssues) == 0, WorkloadsReady: true, ConfiguredProbes: true, StatefulComplete: statefulComplete, ParityDigest: observed.ParityDigest}
	for _, item := range observed.Items {
		if item.Health != nil && !strings.EqualFold(item.Health.Status, "ready") {
			r.WorkloadsReady = false
		}
	}
	if !observed.Complete {
		r.Reasons = append(r.Reasons, "feature inventory is incomplete")
	}
	if !r.BackendApplyComplete {
		r.Reasons = append(r.Reasons, "backend apply is incomplete")
	}
	if !r.StructuralParity {
		r.Reasons = append(r.Reasons, "feature inventory differs from applied release plan")
	}
	if !r.NoForbiddenBaseLinks {
		r.Reasons = append(r.Reasons, "feature has forbidden base links")
	}
	if !r.WorkloadsReady {
		r.Reasons = append(r.Reasons, "feature workloads are not ready")
	}
	if !r.ConfiguredProbes {
		r.Reasons = append(r.Reasons, "configured probes are missing")
	}
	if !r.StatefulComplete {
		r.Reasons = append(r.Reasons, "stateful seed or migration is incomplete")
	}
	r.Ready = observed.Complete && r.BackendApplyComplete && r.StructuralParity && r.NoForbiddenBaseLinks && r.WorkloadsReady && r.ConfiguredProbes && r.StatefulComplete
	return r
}

func inventoryIdentity(item ReleasePlanInventoryItem) string {
	return item.Kind + "/" + item.Namespace + "/" + item.Name
}
func inventoryByIdentity(items []ReleasePlanInventoryItem) map[string]ReleasePlanInventoryItem {
	out := map[string]ReleasePlanInventoryItem{}
	for _, item := range items {
		out[inventoryIdentity(item)] = item
	}
	return out
}
func sortInventoryItems(items []ReleasePlanInventoryItem) {
	sort.Slice(items, func(i, j int) bool { return inventoryIdentity(items[i]) < inventoryIdentity(items[j]) })
}
