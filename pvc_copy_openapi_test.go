package contracts

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/envplane/contracts/domain"
)

func TestPVCCopyCanonicalOpenAPIFields(t *testing.T) {
	raw, err := os.ReadFile("openapi/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	type schema struct {
		Properties           map[string]json.RawMessage `json:"properties"`
		Required             []string                   `json:"required"`
		AdditionalProperties *bool                      `json:"additionalProperties"`
	}
	var doc struct {
		Components struct {
			Schemas map[string]schema `json:"schemas"`
		} `json:"components"`
		Paths map[string]struct {
			Post struct {
				RequestBody struct {
					Content map[string]struct {
						Schema schema `json:"schema"`
					} `json:"content"`
				} `json:"requestBody"`
			} `json:"post"`
		} `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	for name, typ := range map[string]reflect.Type{"PVCCopyPlan": reflect.TypeOf(domain.PVCCopyPlan{}), "PVCCopyItem": reflect.TypeOf(domain.PVCCopyItem{}), "MySQLRestorePlan": reflect.TypeOf(domain.MySQLRestorePlan{}), "MySQLRestoreItem": reflect.TypeOf(domain.MySQLRestoreItem{}), "MySQLRestoreSource": reflect.TypeOf(domain.MySQLRestoreSource{}), "MySQLRestoreResult": reflect.TypeOf(domain.MySQLRestoreResult{}), "MySQLRestoreTLS": reflect.TypeOf(domain.MySQLRestoreTLS{}), "MySQLRestoreCredentialRef": reflect.TypeOf(domain.MySQLRestoreCredentialRef{})} {
		s := doc.Components.Schemas[name]
		if s.AdditionalProperties == nil || *s.AdditionalProperties {
			t.Fatalf("%s must reject unknown fields", name)
		}
		if len(s.Properties) != typ.NumField() {
			t.Fatalf("%s schema field count differs from domain", name)
		}
		for i := 0; i < typ.NumField(); i++ {
			field := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
			if _, ok := s.Properties[field]; !ok {
				t.Fatalf("%s missing property %s", name, field)
			}
			found := false
			for _, required := range s.Required {
				if required == field {
					found = true
				}
			}
			optional := strings.Contains(typ.Field(i).Tag.Get("json"), ",omitempty")
			if found == optional {
				t.Fatalf("%s missing required field %s", name, field)
			}
		}
	}
	for name, fields := range map[string][]string{
		"RunnerCommand": {"pvcCopyPlan", "mysqlRestorePlan"}, "RunnerCommandResult": {"pvcCopyPlanDigest", "pvcCopyVerified", "mysqlRestorePlanDigest", "mysqlRestoreVerified"},
		"ResourceSnapshot": {"sourceUid"}, "RunnerHeartbeatRequest": {"pvcCopyContractVersion", "mysqlRestoreContractVersion", "mysqlRestoreSourceNamespaces", "mysqlRestoreTargetImages", "mysqlRestoreHelperImage"},
		"SecretMaterializationItemResult": {"outputUid"},
	} {
		for _, field := range fields {
			if _, ok := doc.Components.Schemas[name].Properties[field]; !ok {
				t.Fatalf("%s missing %s", name, field)
			}
		}
	}
	if _, ok := doc.Paths["/api/v1/runners/heartbeat"].Post.RequestBody.Content["application/json"].Schema.Properties["pvcCopyContractVersion"]; !ok {
		t.Fatal("heartbeat endpoint missing capability metadata")
	}
	for _, field := range []string{"mysqlRestoreContractVersion", "mysqlRestoreSourceNamespaces", "mysqlRestoreTargetImages", "mysqlRestoreHelperImage"} {
		if _, ok := doc.Paths["/api/v1/runners/heartbeat"].Post.RequestBody.Content["application/json"].Schema.Properties[field]; !ok {
			t.Fatalf("heartbeat endpoint missing %s", field)
		}
	}
}
