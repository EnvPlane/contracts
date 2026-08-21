// Package envplanesdk is the public, versioned integration SDK. It contains
// transport and contract helpers only; Enterprise implementations stay private.
package envplanesdk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	Headers       http.Header
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
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, values := range c.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
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

// DoJSON performs a typed JSON request while retaining the common SDK
// authentication, timeout and idempotency behavior.
func (c Client) DoJSON(ctx context.Context, method, path string, request any, response any, idempotencyKey string) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("encode SDK request: %w", err)
	}
	resp, err := c.Do(ctx, method, path, bytes.NewReader(body), idempotencyKey)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var apiError APIError
		if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&apiError); err != nil {
			apiError = APIError{Status: resp.StatusCode}
		}
		apiError.Status = resp.StatusCode
		return &apiError
	}
	if response == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
		return fmt.Errorf("decode SDK response: %w", err)
	}
	return nil
}

type Page struct {
	Limit  int
	Cursor string
}

func (p Page) Encode() string { return p.Values().Encode() }

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

// NextPage reads the opaque cursor emitted by the API without interpreting it.
func NextPage(r *http.Response) (Page, bool) {
	if r == nil {
		return Page{}, false
	}
	cursor := strings.TrimSpace(r.Header.Get("X-EnvPlane-Next-Cursor"))
	if cursor == "" {
		return Page{}, false
	}
	return Page{Cursor: cursor}, true
}

type APIError struct {
	Status     int    `json:"-"`
	Code       string `json:"code,omitempty"`
	Feature    string `json:"feature,omitempty"`
	Limit      string `json:"limit,omitempty"`
	Current    int64  `json:"current,omitempty"`
	Requested  int64  `json:"requested,omitempty"`
	Plan       string `json:"plan,omitempty"`
	UpgradeURL string `json:"upgrade_url,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
}

func (e *APIError) Error() string {
	if e == nil {
		return "EnvPlane API request failed"
	}
	if e.Code != "" {
		return fmt.Sprintf("EnvPlane API request failed: status=%d code=%s", e.Status, e.Code)
	}
	return fmt.Sprintf("EnvPlane API request failed: status=%d", e.Status)
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
	out := make([]string, 0, min(len(parts), 64))
	seen := make(map[string]struct{}, len(out))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			if len(out) >= 64 {
				break
			}
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			out = append(out, value)
		}
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
