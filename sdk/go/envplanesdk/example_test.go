package envplanesdk_test

import (
	"context"
	"net/http"

	"github.com/envpilot/contracts/sdk/go/envplanesdk"
)

// ExampleClient demonstrates runtime-only credentials; no token is persisted.
func ExampleClient() {
	client := envplanesdk.Client{BaseURL: "https://control-plane.example", TokenProvider: func(context.Context) (string, error) { return "runtime-token-from-secret-store", nil }}
	_, _ = client.Do(context.Background(), http.MethodGet, "/api/v1/projects", nil, "projects-read-1")
}
