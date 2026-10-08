package domain

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

type BaseResourcePin struct {
	Namespace    string `json:"namespace"`
	ResourceKind string `json:"resourceKind"`
	ResourceName string `json:"resourceName"`
	ResourceUID  string `json:"resourceUid"`
	ComponentID  string `json:"componentId"`
}

type BaseResourceBinding struct {
	BaseResourcePin
	ID                string     `json:"id"`
	TenantID          string     `json:"tenantId"`
	ProjectID         string     `json:"projectId"`
	ClusterID         string     `json:"clusterId"`
	ClusterGeneration int64      `json:"clusterGeneration"`
	Version           int64      `json:"version"`
	State             string     `json:"state"`
	CreatedAt         time.Time  `json:"createdAt"`
	CreatedBy         string     `json:"createdBy"`
	RevokedAt         *time.Time `json:"revokedAt,omitempty"`
	RevokedBy         string     `json:"revokedBy,omitempty"`
}

// No authoritative tenant, actor or generation is accepted from request bodies.
type RegisterBaseResourceBinding struct {
	BaseResourcePin
	ProjectID      string `json:"projectId"`
	ClusterID      string `json:"clusterId"`
	IdempotencyKey string `json:"idempotencyKey"`
}

type FinOpsResourceAttribution struct {
	EnvironmentID         string `json:"environmentId,omitempty"`
	BaseResourceBindingID string `json:"baseResourceBindingId,omitempty"`
	BindingVersion        int64  `json:"bindingVersion,omitempty"`
}

func (a FinOpsResourceAttribution) Validate() error {
	if (a.EnvironmentID == "") == (a.BaseResourceBindingID == "") || (a.EnvironmentID != "" && a.BindingVersion != 0) || (a.BaseResourceBindingID != "" && a.BindingVersion < 1) {
		return errors.New("exactly one environment or versioned baseline binding required")
	}
	return nil
}

var baseResourceName = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]*$`)
var baseResourceUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func (p BaseResourcePin) Validate() error {
	if (p.ResourceKind != "Pod" && p.ResourceKind != "PersistentVolumeClaim") || len(p.Namespace) > 63 || len(p.ResourceName) > 253 || !baseResourceName.MatchString(p.Namespace) || !baseResourceName.MatchString(p.ResourceName) || !baseResourceUID.MatchString(p.ResourceUID) || p.ComponentID == "" || len(p.ComponentID) > 128 || strings.TrimSpace(p.ComponentID) != p.ComponentID {
		return errors.New("explicit valid resource pin required")
	}
	return nil
}

type BaselineMeteringSample struct {
	Attribution FinOpsResourceAttribution `json:"attribution"`
	BaseResourcePin
	SampleID        string  `json:"sampleId"`
	Metric          string  `json:"metric"`
	Unit            string  `json:"unit"`
	MeasurementKind string  `json:"measurementKind"`
	Source          string  `json:"source"`
	Quantity        float64 `json:"quantity"`
	UsedBytes       *int64  `json:"usedBytes,omitempty"`
}

type BaselineMeteringBatch struct {
	BatchID     string                   `json:"batchId"`
	ProjectID   string                   `json:"projectId"`
	ClusterID   string                   `json:"clusterId"`
	AgentID     string                   `json:"agentId"`
	PeriodStart time.Time                `json:"periodStart"`
	PeriodEnd   time.Time                `json:"periodEnd"`
	Samples     []BaselineMeteringSample `json:"samples"`
}
