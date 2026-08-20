package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/envplane/contracts/sdk/go/envplanesdk"
)

func TestCanonicalOpenAPIBrandingAndGeneratedSDKStayInSync(t *testing.T) {
	raw, err := os.ReadFile("openapi/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Info struct {
			Title string `json:"title"`
			Extensions map[string]any `json:"x-branding"`
		} `json:"info"`
	}
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	if document.Info.Title != "EnvPlane API" || document.Info.Extensions["canonicalProduct"] != "EnvPlane" {
		t.Fatalf("unexpected canonical OpenAPI identity: %#v", document.Info)
	}
	digest := sha256.Sum256(raw)
	if got := hex.EncodeToString(digest[:]); got != envplanesdk.CanonicalOpenAPISHA256 {
		t.Fatalf("generated SDK hash is stale: got %s want %s; run go generate ./sdk/go/envplanesdk", envplanesdk.CanonicalOpenAPISHA256, got)
	}
}
