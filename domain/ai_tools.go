package domain

import (
	"errors"
	"strings"
)

const AIToolRegistrySchemaVersion = "1"

type AIToolDescriptor struct {
	SchemaVersion  string       `json:"schemaVersion"`
	Name           string       `json:"name"`
	Risk           AIRiskClass  `json:"risk"`
	ActionClass    AIActionClass `json:"actionClass"`
	Permissions    []string     `json:"permissions"`
	RequiredArgs   []string     `json:"requiredArgs"`
	AllowedArgs    []string     `json:"allowedArgs"`
	ReadOnly       bool         `json:"readOnly"`
	SupportsDryRun bool         `json:"supportsDryRun"`
	Compensation   string       `json:"compensation,omitempty"`
	Preconditions  []string     `json:"preconditions"`
}

type AIToolObservation struct {
	SchemaVersion string            `json:"schemaVersion"`
	Code          string            `json:"code"`
	Fields        map[string]string `json:"fields"`
}

func (d AIToolDescriptor) Validate() error {
	if d.SchemaVersion != AIToolRegistrySchemaVersion || strings.TrimSpace(d.Name) == "" || len(d.Name) > 128 {
		return errors.New("AI tool descriptor identity is invalid")
	}
	switch d.Risk {
	case AIRiskLow, AIRiskMedium, AIRiskHigh, AIRiskCritical:
	default:
		return errors.New("AI tool descriptor risk is invalid")
	}
	switch d.ActionClass {
	case AIActionReadOnly, AIActionProposal, AIActionApprovedWrite, AIActionForbidden:
	default:
		return errors.New("AI tool descriptor action class is invalid")
	}
	if d.ActionClass == AIActionForbidden || (d.ReadOnly && d.ActionClass == AIActionApprovedWrite) {
		return errors.New("AI tool descriptor action class conflicts with read-only policy")
	}
	return nil
}

func (o AIToolObservation) Validate() error {
	if o.SchemaVersion != AIToolRegistrySchemaVersion || strings.TrimSpace(o.Code) == "" || len(o.Code) > 128 || len(o.Fields) > 32 {
		return errors.New("AI tool observation is invalid or unbounded")
	}
	for key, value := range o.Fields {
		if strings.TrimSpace(key) == "" || len(key) > 128 || len(value) > 1024 {
			return errors.New("AI tool observation field is invalid or unbounded")
		}
	}
	return nil
}
