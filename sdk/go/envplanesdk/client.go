// Package envplanesdk is the public, versioned integration SDK. It contains
// transport and contract helpers only; Enterprise implementations stay private.
package envplanesdk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const Version = "v1.0.0"

type TokenProvider func(context.Context) (string, error)

type Client struct {
	BaseURL       string
	HTTPClient    *http.Client
	TokenProvider TokenProvider
}

func (c Client) Do(ctx context.Context, method, path string, body io.Reader, idempotencyKey string) (*http.Response, error) {
	base, err := url.Parse(strings.TrimRight(c.BaseURL, "/"))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid SDK base URL")
	}
	rel, err := url.Parse("/" + strings.TrimLeft(path, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid SDK path: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, base.ResolveReference(rel).String(), body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "envplane-sdk/"+Version)
	if strings.TrimSpace(idempotencyKey) != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if c.TokenProvider != nil {
		token, err := c.TokenProvider(ctx)
		if err != nil {
			return nil, fmt.Errorf("get SDK token: %w", err)
		}
		if strings.TrimSpace(token) == "" {
			return nil, fmt.Errorf("SDK token is empty")
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}
	return client.Do(req)
}

type Page struct {
	Limit  int
	Cursor string
}

func (p Page) Values() url.Values {
	v := url.Values{}
	if p.Limit > 0 {
		v.Set("limit", strconv.Itoa(p.Limit))
	}
	if strings.TrimSpace(p.Cursor) != "" {
		v.Set("cursor", p.Cursor)
	}
	return v
}

type Capabilities struct {
	APIVersion string   `json:"apiVersion"`
	Features   []string `json:"features"`
}

func NegotiateCapabilities(r *http.Response) Capabilities {
	if r == nil {
		return Capabilities{}
	}
	return Capabilities{APIVersion: r.Header.Get("X-EnvPlane-API-Version"), Features: splitHeader(r.Header.Get("X-EnvPlane-Capabilities"))}
}

func splitHeader(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func VerifyWebhook(secret, signature string, payload []byte) bool {
	if strings.TrimSpace(secret) == "" || strings.TrimSpace(signature) == "" {
		return false
	}
	provided := strings.TrimPrefix(strings.TrimSpace(signature), "sha256=")
	decoded, err := hex.DecodeString(provided)
	if err != nil {
		return false
	}
	digest := hmac.New(sha256.New, []byte(secret))
	_, _ = digest.Write(payload)
	return hmac.Equal(decoded, digest.Sum(nil))
}
