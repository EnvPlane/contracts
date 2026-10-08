package domain

import (
	"errors"
	"regexp"
	"time"
)

const NetworkPolicyProbeSchemaVersion = 1
const NetworkPolicyProbeFreshness = 15 * time.Minute
const NetworkPolicyProbeInternalKey = "__envplane_network_policy_probe"

var ErrNetworkPolicyProbeInvalid = errors.New("invalid network policy probe evidence")
var probeClusterUIDPattern = regexp.MustCompile(`^[a-zA-Z0-9-]{1,64}$`)

func ValidNetworkPolicyProbeClusterUID(value string) bool {
	return probeClusterUIDPattern.MatchString(value)
}

type NetworkPolicyProbeReport struct {
	SchemaVersion   int       `json:"schemaVersion"`
	ClusterUID      string    `json:"clusterUID"`
	Generation      int64     `json:"generation"`
	CheckedAt       time.Time `json:"checkedAt"`
	Scope           string    `json:"scope"`
	Ingress         string    `json:"ingress"`
	Egress          string    `json:"egress"`
	State           string    `json:"state"`
	Reason          string    `json:"reason"`
	CleanupComplete bool      `json:"cleanupComplete"`
}

func (p NetworkPolicyProbeReport) Validate() error {
	if p.SchemaVersion != NetworkPolicyProbeSchemaVersion || !probeClusterUIDPattern.MatchString(p.ClusterUID) || p.Generation <= 0 || p.CheckedAt.IsZero() || p.Scope != "single-node-pod-ipv4-tcp" {
		return ErrNetworkPolicyProbeInvalid
	}
	valid := func(v string) bool { return v == "passed" || v == "failed" || v == "unknown" }
	if !valid(p.State) || !valid(p.Ingress) || !valid(p.Egress) {
		return ErrNetworkPolicyProbeInvalid
	}
	switch p.Reason {
	case "measured", "policy_not_enforced", "probe_unavailable", "cleanup_failed", "baseline_failed", "recovery_failed", "cluster_identity_changed":
	default:
		return ErrNetworkPolicyProbeInvalid
	}
	if p.State == "passed" && (p.Ingress != "passed" || p.Egress != "passed" || !p.CleanupComplete || p.Reason != "measured") {
		return ErrNetworkPolicyProbeInvalid
	}
	if p.State == "failed" && (p.Ingress != "failed" && p.Egress != "failed" || !p.CleanupComplete || p.Reason != "policy_not_enforced") {
		return ErrNetworkPolicyProbeInvalid
	}
	return nil
}

type NetworkPolicyProbeIdentity struct {
	ProjectID string `json:"projectId"`
	ClusterID string `json:"clusterId"`
	AgentID   string `json:"agentId"`
}
type NetworkPolicyProbeChallengeRequest struct {
	NetworkPolicyProbeIdentity
	ClusterUID string `json:"clusterUID"`
}
type NetworkPolicyProbeChallenge struct {
	ProbeID    string    `json:"probeId"`
	Nonce      string    `json:"nonce"`
	Generation int64     `json:"generation"`
	IssuedAt   time.Time `json:"issuedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
}
type NetworkPolicyProbeSubmission struct {
	NetworkPolicyProbeIdentity
	ProbeID string                   `json:"probeId"`
	Nonce   string                   `json:"nonce"`
	Report  NetworkPolicyProbeReport `json:"report"`
}
type NetworkPolicyProbeObservation struct {
	Report     *NetworkPolicyProbeReport `json:"report,omitempty"`
	ReceivedAt *time.Time                `json:"receivedAt,omitempty"`
	Fresh      bool                      `json:"fresh"`
	State      string                    `json:"state"`
}
