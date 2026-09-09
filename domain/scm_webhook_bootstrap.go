package domain

import "time"

// SCMWebhookBootstrapProof is safe to return to browser clients. It records
// only public webhook metadata and one-way fingerprints, never credentials.
type SCMWebhookBootstrapProof struct {
	Provider          string    `json:"provider"`
	RepositoryID      string    `json:"repositoryId"`
	RepositoryPath    string    `json:"repositoryPath"`
	HookID            int64     `json:"hookId"`
	CallbackURL       string    `json:"callbackUrl"`
	ConfigFingerprint string    `json:"configFingerprint"`
	SecretFingerprint string    `json:"secretFingerprint"`
	EndpointState     string    `json:"endpointState"`
	DNSState          string    `json:"dnsState"`
	TLSState          string    `json:"tlsState"`
	ReceiverState     string    `json:"receiverState"`
	DeliveryState     string    `json:"deliveryState"`
	FailureCode       string    `json:"failureCode"`
	ObservedAt        time.Time `json:"observedAt"`
	VerifiedAt        time.Time `json:"verifiedAt"`
}

// Ready is the single readiness predicate used by Bootstrap gates.
func (p SCMWebhookBootstrapProof) Ready() bool {
	return p.EndpointState == "ready" && p.ReceiverState == "ready" && p.DeliveryState == "verified"
}

// ReadyForConfig additionally binds a verified proof to the currently desired
// webhook configuration. An empty expected fingerprint is never ready.
func (p SCMWebhookBootstrapProof) ReadyForConfig(expectedFingerprint string) bool {
	return expectedFingerprint != "" && p.ConfigFingerprint == expectedFingerprint && p.Ready()
}
