// Command example demonstrates a credential-free integration composition.
// The token is read only at runtime from an operator-managed environment.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/envplane/contracts/sdk/go/envplanesdk"
)

func main() {
	client := envplanesdk.Client{
		BaseURL: os.Getenv("ENVPLANE_API_URL"),
		TokenProvider: func(context.Context) (string, error) {
			return os.Getenv("ENVPLANE_API_TOKEN"), nil
		},
	}
	response, err := client.Do(context.Background(), http.MethodGet, "/api/v1/capabilities", nil, "")
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	capabilities := envplanesdk.NegotiateCapabilities(response)
	fmt.Printf("status=%d api=%s features=%v\n", response.StatusCode, capabilities.APIVersion, capabilities.Features)
}
