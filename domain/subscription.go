package domain

import (
	"fmt"
	"time"
)

type SubscriptionState string

const (
	SubscriptionTrialing   SubscriptionState = "trialing"
	SubscriptionActive     SubscriptionState = "active"
	SubscriptionPastDue    SubscriptionState = "past_due"
	SubscriptionGrace      SubscriptionState = "grace"
	SubscriptionCanceled   SubscriptionState = "canceled"
	SubscriptionDowngraded SubscriptionState = "downgraded"
)

type Subscription struct {
	TenantID        string            `json:"tenantId"`
	ID              string            `json:"id"`
	PlanID          string            `json:"planId"`
	PlanVersion     string            `json:"planVersion"`
	Provider        string            `json:"provider,omitempty"`
	ProviderRef     string            `json:"providerRef,omitempty"`
	ProviderEventID string            `json:"providerEventId,omitempty"`
	State           SubscriptionState `json:"state"`
	PeriodStart     time.Time         `json:"periodStart"`
	PeriodEnd       time.Time         `json:"periodEnd"`
	GraceEnd        *time.Time        `json:"graceEnd,omitempty"`
	Source          string            `json:"source"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
}

type Entitlement struct {
	TenantID             string           `json:"tenantId"`
	PlanID               string           `json:"planId"`
	PlanVersion          string           `json:"planVersion"`
	Features             map[string]bool  `json:"features"`
	Limits               map[string]int64 `json:"limits"`
	SourceSubscriptionID string           `json:"sourceSubscriptionId"`
	UpdatedAt            time.Time        `json:"updatedAt"`
}

func ValidateSubscriptionTransition(from, to SubscriptionState) error {
	if from != "" && !validSubscriptionState(from) {
		return fmt.Errorf("invalid current subscription state %q", from)
	}
	if !validSubscriptionState(to) {
		return fmt.Errorf("invalid subscription state %q", to)
	}
	if from == "" {
		if to == SubscriptionTrialing || to == SubscriptionActive {
			return nil
		}
		return fmt.Errorf("invalid initial subscription state %q", to)
	}
	if from == to {
		return nil
	}
	allowed := map[SubscriptionState]map[SubscriptionState]bool{
		SubscriptionTrialing:   {SubscriptionActive: true, SubscriptionGrace: true, SubscriptionCanceled: true},
		SubscriptionActive:     {SubscriptionPastDue: true, SubscriptionGrace: true, SubscriptionCanceled: true},
		SubscriptionPastDue:    {SubscriptionGrace: true, SubscriptionActive: true, SubscriptionCanceled: true},
		SubscriptionGrace:      {SubscriptionActive: true, SubscriptionCanceled: true, SubscriptionDowngraded: true},
		SubscriptionCanceled:   {},
		SubscriptionDowngraded: {SubscriptionActive: true, SubscriptionCanceled: true},
	}
	if !allowed[from][to] {
		return fmt.Errorf("invalid subscription transition %q -> %q", from, to)
	}
	return nil
}

func validSubscriptionState(state SubscriptionState) bool {
	switch state {
	case SubscriptionTrialing, SubscriptionActive, SubscriptionPastDue, SubscriptionGrace, SubscriptionCanceled, SubscriptionDowngraded:
		return true
	default:
		return false
	}
}
