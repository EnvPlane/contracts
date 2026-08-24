package domain

import (
	"errors"
	"strings"
	"time"
)

const AIFinOpsOptimizationSchemaVersion = "1"

type AIFinOpsOptimizationKind string

const (
	AIFinOpsOptimizeTTL            AIFinOpsOptimizationKind = "ttl"
	AIFinOpsOptimizeSchedule       AIFinOpsOptimizationKind = "schedule"
	AIFinOpsOptimizeRequestsLimits AIFinOpsOptimizationKind = "requests_limits"
	AIFinOpsOptimizeSharedService  AIFinOpsOptimizationKind = "shared_service"
)

type AIFinOpsConfidence struct {
	Level         string  `json:"level"`
	Score         float64 `json:"score"`
	EvidenceCount int     `json:"evidenceCount"`
}

type AIFinOpsOptimizationProposal struct {
	ID                   string                      `json:"id"`
	Kind                 AIFinOpsOptimizationKind    `json:"kind"`
	ProposalFields       ConfigurationProposalFields `json:"proposalFields"`
	Savings              AIFinOpsMoney               `json:"savings"`
	Confidence           AIFinOpsConfidence          `json:"confidence"`
	Risk                 string                      `json:"risk"`
	PerformanceTradeoff  string                      `json:"performanceTradeoff"`
	Evidence             []AIEvidenceReference       `json:"evidence"`
	ApprovalRequired     bool                        `json:"approvalRequired"`
	PostVerification     []string                    `json:"postVerification"`
	BillingAuthoritative bool                        `json:"billingAuthoritative"`
}

type AIFinOpsOptimizationPlan struct {
	SchemaVersion string                         `json:"schemaVersion"`
	TenantID      string                         `json:"tenantId"`
	ProjectID     string                         `json:"projectId"`
	Currency      string                         `json:"currency"`
	PeriodStart   time.Time                      `json:"periodStart"`
	PeriodEnd     time.Time                      `json:"periodEnd"`
	Forecast      AIFinOpsMoney                  `json:"forecast"`
	Confidence    AIFinOpsConfidence             `json:"confidence"`
	Evidence      []AIEvidenceReference          `json:"evidence"`
	Proposals     []AIFinOpsOptimizationProposal `json:"proposals,omitempty"`
	MissingData   []string                       `json:"missingData,omitempty"`
	GeneratedAt   time.Time                      `json:"generatedAt"`
}

func (p AIFinOpsOptimizationPlan) Validate() error {
	if p.SchemaVersion != AIFinOpsOptimizationSchemaVersion || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" || strings.TrimSpace(p.Currency) == "" || p.PeriodStart.Location() != time.UTC || p.PeriodEnd.Location() != time.UTC || !p.PeriodEnd.After(p.PeriodStart) || p.GeneratedAt.IsZero() {
		return errors.New("invalid FinOps optimization plan")
	}
	if p.Forecast.Currency != p.Currency {
		return errors.New("optimization forecast currency mismatch")
	}
	for _, evidence := range p.Evidence {
		if evidence.TenantID != p.TenantID || evidence.Validate() != nil {
			return errors.New("optimization evidence is outside tenant scope")
		}
	}
	for _, proposal := range p.Proposals {
		if strings.TrimSpace(proposal.ID) == "" || proposal.ProposalFields.Validate() != nil || len(proposal.PostVerification) == 0 || !proposal.BillingAuthoritative {
			return errors.New("optimization proposal is not safely typed")
		}
		if proposal.Savings.Known {
			if proposal.Savings.Currency != p.Currency || proposal.Savings.MinorUnits < 0 || len(proposal.Evidence) == 0 {
				return errors.New("known savings require grounded non-negative evidence")
			}
		}
		for _, evidence := range proposal.Evidence {
			if evidence.TenantID != p.TenantID || evidence.Validate() != nil {
				return errors.New("proposal evidence is outside tenant scope")
			}
		}
		if proposal.Kind == AIFinOpsOptimizeRequestsLimits || proposal.Kind == AIFinOpsOptimizeSharedService {
			if !proposal.ApprovalRequired {
				return errors.New("resource or shared-service optimization requires approval")
			}
		}
	}
	return nil
}
