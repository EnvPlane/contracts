package domain

import "time"

type BillingCustomerRequest struct {
	TenantID string `json:"tenantId"`
}

type BillingCheckoutRequest struct {
	TenantID       string `json:"tenantId"`
	PlanID         string `json:"planId"`
	PlanVersion    string `json:"planVersion"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type BillingCheckoutSession struct {
	TenantID   string    `json:"tenantId"`
	SessionRef string    `json:"sessionRef"`
	PlanID     string    `json:"planId"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type BillingPortalSession struct {
	TenantID   string    `json:"tenantId"`
	SessionRef string    `json:"sessionRef"`
	ExpiresAt  time.Time `json:"expiresAt"`
}

type BillingUsageReport struct {
	TenantID       string    `json:"tenantId"`
	Metric         string    `json:"metric"`
	Quantity       int64     `json:"quantity"`
	PeriodStart    time.Time `json:"periodStart"`
	PeriodEnd      time.Time `json:"periodEnd"`
	IdempotencyKey string    `json:"idempotencyKey"`
}
