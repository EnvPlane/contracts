package domain

import "time"

// SCMWebhookBootstrapProof is safe to return to browser clients. It records
// only public webhook metadata and one-way fingerprints, never credentials.
type SCMWebhookBootstrapProof struct {
	Provider                  string    `json:"provider"`
	RepositoryID              string    `json:"repositoryId"`
	RepositoryPath            string    `json:"repositoryPath"`
	HookID                    int64     `json:"hookId"`
	CallbackURL               string    `json:"callbackUrl"`
	ConfigFingerprint         string    `json:"configFingerprint"`
	SecretFingerprint         string    `json:"secretFingerprint"`
	PreviousSecretFingerprint string    `json:"previousSecretFingerprint,omitempty"`
	PreviousSecretExpiresAt   time.Time `json:"previousSecretExpiresAt,omitempty"`
	HookState                 string    `json:"hookState,omitempty"`
	LastDriftCheckAt          time.Time `json:"lastDriftCheckAt,omitempty"`
	EndpointState             string    `json:"endpointState"`
	DNSState                  string    `json:"dnsState"`
	TLSState                  string    `json:"tlsState"`
	ReceiverState             string    `json:"receiverState"`
	DeliveryState             string    `json:"deliveryState"`
	DeliveryNonce             string    `json:"deliveryNonce,omitempty"`
	DeliveryDeadlineAt        time.Time `json:"deliveryDeadlineAt,omitempty"`
	FailureCode               string    `json:"failureCode"`
	ObservedAt                time.Time `json:"observedAt"`
	VerifiedAt                time.Time `json:"verifiedAt"`
	LastDeliveryAt            time.Time `json:"lastDeliveryAt,omitempty"`
	LastSignatureResult       string    `json:"lastSignatureResult,omitempty"`
	LastJobID                 string    `json:"lastJobId,omitempty"`
	LastJobStatus             string    `json:"lastJobStatus,omitempty"`
}

// SCMWebhookStatus is the redacted operational read model used by project
// settings and support diagnostics. It never contains webhook credentials,
// GitLab tokens, or event payloads.
type SCMWebhookStatus struct {
	Provider            string    `json:"provider"`
	PublicURL           string    `json:"publicUrl,omitempty"`
	EndpointState       string    `json:"endpointState"`
	DNSState            string    `json:"dnsState"`
	TLSState            string    `json:"tlsState"`
	ReceiverState       string    `json:"receiverState"`
	ReceiverHealth      string    `json:"receiverHealth"`
	LastDeliveryAt      time.Time `json:"lastDeliveryAt,omitempty"`
	LastDeliveryState   string    `json:"lastDeliveryState"`
	LastSignatureResult string    `json:"lastSignatureResult"`
	LastJobID           string    `json:"lastJobId,omitempty"`
	LastJobStatus       string    `json:"lastJobStatus,omitempty"`
	FailureCode         string    `json:"failureCode,omitempty"`
	ObservedAt          time.Time `json:"observedAt"`
	Ready               bool      `json:"ready"`
	ReadinessBlocker    string    `json:"readinessBlocker,omitempty"`
}

// SCMWebhookMigrationPreflight is a read-only, redacted migration report.
// It intentionally exposes no Secret data or signing credentials.
type SCMWebhookMigrationPreflight struct {
	Provider             string    `json:"provider"`
	ProjectID            string    `json:"projectId"`
	CredentialConfigured bool      `json:"credentialConfigured"`
	ProofConfigured      bool      `json:"proofConfigured"`
	CallbackURL          string    `json:"callbackUrl,omitempty"`
	EndpointState        string    `json:"endpointState"`
	DNSState             string    `json:"dnsState"`
	TLSState             string    `json:"tlsState"`
	ReceiverState        string    `json:"receiverState"`
	DeliveryState        string    `json:"deliveryState"`
	LegacyFallbackMode   string    `json:"legacyFallbackMode"`
	LegacyFallbackUntil  time.Time `json:"legacyFallbackUntil,omitempty"`
	ReadOnly             bool      `json:"readOnly"`
	RecommendedAction    string    `json:"recommendedAction,omitempty"`
}

func SCMWebhookStatusFromProof(p SCMWebhookBootstrapProof) SCMWebhookStatus {
	status := SCMWebhookStatus{
		Provider: p.Provider, PublicURL: p.CallbackURL, EndpointState: p.EndpointState,
		DNSState: p.DNSState, TLSState: p.TLSState, ReceiverState: p.ReceiverState,
		ReceiverHealth: p.ReceiverState, LastDeliveryAt: p.LastDeliveryAt,
		LastDeliveryState: p.DeliveryState, LastSignatureResult: p.LastSignatureResult,
		LastJobID: p.LastJobID, LastJobStatus: p.LastJobStatus, FailureCode: p.FailureCode,
		ObservedAt: p.ObservedAt,
	}
	status.Ready = p.Ready()
	if !status.Ready {
		status.ReadinessBlocker = p.FailureCode
		if status.ReadinessBlocker == "" {
			status.ReadinessBlocker = "webhook_delivery_not_verified"
		}
	}
	return status
}

// Ready is the single readiness predicate used by Bootstrap gates.
func (p SCMWebhookBootstrapProof) Ready() bool {
	return p.EndpointState == "ready" && p.ReceiverState == "ready" && p.DeliveryState == "verified" && (p.HookState == "" || p.HookState == "ready")
}

// ReadyForConfig additionally binds a verified proof to the currently desired
// webhook configuration. An empty expected fingerprint is never ready.
func (p SCMWebhookBootstrapProof) ReadyForConfig(expectedFingerprint string) bool {
	return expectedFingerprint != "" && p.ConfigFingerprint == expectedFingerprint && p.Ready()
}
