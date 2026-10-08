package domain

// Commands are durable seller-side intent, not proof of payment, activation or
// provider completion. Installation scope never comes from browser request data.
type BillingCommandRequest struct {
	CommandID        string `json:"commandId"`
	ExpectedRevision int64  `json:"expectedRevision"`
	Kind             string `json:"kind"`
	PlanID           string `json:"planId,omitempty"`
}
type BillingCommandResult struct {
	CommandID string `json:"commandId"`
	Revision  int64  `json:"revision"`
	Status    string `json:"status"`
}
type BillingCommandState struct {
	TenantID  string                 `json:"tenantId"`
	Revision  int64                  `json:"revision"`
	Available bool                   `json:"available"`
	Commands  []BillingCommandResult `json:"commands"`
}
