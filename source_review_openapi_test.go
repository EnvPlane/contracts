package contracts

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/envplane/contracts/sdk/go/envplanesdk"
)

func sourceReviewDocument(t *testing.T) map[string]any {
	t.Helper()
	data, err := os.ReadFile("openapi/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func sourceReviewMap(t *testing.T, value any) map[string]any {
	t.Helper()
	out, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected schema object, got %T", value)
	}
	return out
}

func TestPVCSourceReviewOpenAPIAndSDKOperations(t *testing.T) {
	doc := sourceReviewDocument(t)
	paths := sourceReviewMap(t, doc["paths"])
	for _, entry := range []struct{ path, method, status, operation string }{
		{"/api/v1/projects/{id}/pvc-copy/source-profile", "get", "200", "getProjectPVCSourceProfile"},
		{"/api/v1/projects/{id}/pvc-copy/source-profile/review", "post", "201", "reviewProjectPVCSourceProfile"},
	} {
		path := sourceReviewMap(t, paths[entry.path])
		op := sourceReviewMap(t, path[entry.method])
		if op["operationId"] != entry.operation {
			t.Fatal("missing or mismatched operation ID")
		}
		if len(op["security"].([]any)) == 0 {
			t.Fatal("source review route must declare authentication")
		}
		responses := sourceReviewMap(t, op["responses"])
		response := sourceReviewMap(t, responses[entry.status])
		content := sourceReviewMap(t, response["content"])
		media := sourceReviewMap(t, content["application/json"])
		schema := sourceReviewMap(t, media["schema"])
		if schema["$ref"] != "#/components/schemas/PVCSourceProfileResponse" {
			t.Fatal("review response contract differs between GET/POST")
		}
		if _, ok := responses["409"]; !ok {
			t.Fatal("preview pending response missing")
		}
		if entry.method == "post" {
			if _, ok := responses["400"]; !ok {
				t.Fatal("bad choices response missing")
			}
			body := sourceReviewMap(t, op["requestBody"])
			if body["required"] != true {
				t.Fatal("review body must be required")
			}
		}
		found := false
		for _, operation := range envplanesdk.CanonicalOperations {
			if operation == strings.ToUpper(entry.method)+" "+entry.path {
				found = true
			}
		}
		if !found {
			t.Fatal("new source review operation missing from generated SDK")
		}
	}
	schemas := sourceReviewMap(t, sourceReviewMap(t, doc["components"])["schemas"])
	response := sourceReviewMap(t, schemas["PVCSourceProfileResponse"])
	properties := sourceReviewMap(t, response["properties"])
	if sourceReviewMap(t, properties["approval"])["$ref"] != "#/components/schemas/ApprovalRequest" {
		t.Fatal("must reuse existing tenant ApprovalRequest")
	}
	if _, ok := paths["/api/v1/tenants/{tenantID}/approvals/{id}"]; !ok {
		t.Fatal("existing approval API missing")
	}
	for path := range paths {
		if strings.Contains(path, "pvc-copy") && strings.Contains(path, "approval") {
			t.Fatal("new source-specific approval decision API must not be added")
		}
	}
}

func TestPVCSourceReviewBooleanChoicesSchema(t *testing.T) {
	doc := sourceReviewDocument(t)
	schemas := sourceReviewMap(t, sourceReviewMap(t, doc["components"])["schemas"])
	s := sourceReviewMap(t, schemas["PVCSourceReviewChoices"])
	if s["type"] != "object" || s["additionalProperties"] != false || s["nullable"] != false {
		t.Fatal("choices must be exact non-null object")
	}
	if !reflect.DeepEqual(s["required"], []any{"allowRootHelpers", "mysqlRestoreAllowRootInit"}) {
		t.Fatal("both root choices required")
	}
	properties := sourceReviewMap(t, s["properties"])
	if len(properties) != 2 {
		t.Fatal("no source/image/authority inputs may be accepted")
	}
	for _, name := range []string{"allowRootHelpers", "mysqlRestoreAllowRootInit"} {
		field := sourceReviewMap(t, properties[name])
		if field["type"] != "boolean" || field["nullable"] != false {
			t.Fatal("choice must be non-null boolean")
		}
	}
	// Evaluate the declared schema's closed object/required/type constraints.
	valid := func(raw string) bool {
		var v any
		if json.Unmarshal([]byte(raw), &v) != nil {
			return false
		}
		obj, ok := v.(map[string]any)
		if !ok || len(obj) != len(properties) {
			return false
		}
		for name := range properties {
			if _, ok := obj[name].(bool); !ok {
				return false
			}
		}
		return true
	}
	for _, raw := range []string{`{"allowRootHelpers":false,"mysqlRestoreAllowRootInit":false}`, `{"allowRootHelpers":true,"mysqlRestoreAllowRootInit":false}`, `{"allowRootHelpers":false,"mysqlRestoreAllowRootInit":true}`, `{"allowRootHelpers":true,"mysqlRestoreAllowRootInit":true}`} {
		if !valid(raw) {
			t.Fatal("explicit boolean choices rejected")
		}
	}
	for _, raw := range []string{`null`, `[]`, `{}`, `{"allowRootHelpers":false}`, `{"allowRootHelpers":null,"mysqlRestoreAllowRootInit":false}`, `{"allowRootHelpers":false,"mysqlRestoreAllowRootInit":null}`, `{"allowRootHelpers":"false","mysqlRestoreAllowRootInit":false}`, `{"allowRootHelpers":0,"mysqlRestoreAllowRootInit":false}`, `{"allowRootHelpers":false,"mysqlRestoreAllowRootInit":false,"sourceUid":"forged"}`, `{"allowRootHelpers":false,"mysqlRestoreAllowRootInit":false} {}`} {
		if valid(raw) {
			t.Fatalf("schema permits unsafe body %s", raw)
		}
	}
	if !strings.Contains(s["description"].(string), "duplicate") {
		t.Fatal("handler's duplicate-key rejection must be documented")
	}
}

func TestPVCSourceReviewMetadataOnlySchema(t *testing.T) {
	doc := sourceReviewDocument(t)
	schemas := sourceReviewMap(t, sourceReviewMap(t, doc["components"])["schemas"])
	input := sourceReviewMap(t, schemas["PVCSourceReviewInput"])
	if input["readOnly"] != true || input["additionalProperties"] != false {
		t.Fatal("source metadata must be server-owned closed response input")
	}
	for _, name := range []string{"PVCSourceReviewInput", "PVCSourceReviewObject", "PVCSourceReview"} {
		s := sourceReviewMap(t, schemas[name])
		properties := sourceReviewMap(t, s["properties"])
		for _, forbidden := range []string{"data", "stringData", "password", "credentials", "dump", "rows", "kubeconfig", "token"} {
			if _, ok := properties[forbidden]; ok {
				t.Fatalf("%s exposes data/credential payload", name)
			}
		}
	}
	object := sourceReviewMap(t, schemas["PVCSourceReviewObject"])
	properties := sourceReviewMap(t, object["properties"])
	kinds := sourceReviewMap(t, properties["kind"])["enum"].([]any)
	for _, kind := range kinds {
		if kind == "Secret" || kind == "Pod" || kind == "PersistentVolumeClaim" {
			t.Fatal("review artifact schema must only expose RBAC/admission previews")
		}
	}
	if object["additionalProperties"] != false {
		t.Fatal("artifact cannot contain arbitrary payload fields")
	}
}
