package domain

import "context"

// TenantPrincipal is the only caller identity accepted by extension points.
// Implementations must treat TenantID as an isolation boundary and never infer
// it from provider credentials or mutable profile attributes.
type TenantPrincipal struct {
	TenantID string
	Subject  string
	Roles    []string
}

func (p TenantPrincipal) Valid() bool { return p.TenantID != "" && p.Subject != "" }

// IdentityRequest and IdentityResult keep the identity boundary provider
// agnostic. Raw assertions and provider secrets never cross this contract.
type IdentityRequest struct {
	Provider string
	Token    string
}

type IdentityResult struct {
	Principal TenantPrincipal
	User      User
}

// EntitlementExtension resolves the effective tenant capabilities.
type EntitlementExtension interface {
	Resolve(context.Context, TenantPrincipal) (EntitlementSnapshot, error)
}

// IdentityExtension authenticates a provider token into a tenant principal.
type IdentityExtension interface {
	Authenticate(context.Context, string, IdentityRequest) (IdentityResult, error)
}

// PolicyExtension evaluates a typed policy input for a tenant principal.
type PolicyExtension interface {
	Evaluate(context.Context, TenantPrincipal, PolicyInput) (PolicyDecision, error)
}

// AuditSinkExtension exports an already-redacted audit event. A sink failure
// must remain observable without blocking cleanup paths.
type AuditSinkExtension interface {
	Emit(context.Context, TenantPrincipal, SIEMAuditEvent) error
}

// FinOpsExtension allocates usage using a caller-provided price table.
type FinOpsExtension interface {
	Allocate(context.Context, TenantPrincipal, []ResourceUsageSample, ResourcePriceTable) ([]CostAllocation, error)
}
