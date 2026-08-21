package envplanesdk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientAuthPaginationAndIdempotency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer runtime-token" || r.Header.Get("Idempotency-Key") != "request-1" {
			http.Error(w, "headers not propagated", http.StatusBadRequest)
			return
		}
		if r.URL.Query().Get("cursor") != "next" || r.URL.Query().Get("limit") != "10" {
			http.Error(w, "pagination not encoded: "+r.URL.RawQuery, http.StatusBadRequest)
			return
		}
		w.Header().Set("X-EnvPlane-API-Version", "1")
		w.Header().Set("X-EnvPlane-Capabilities", "billing,scim")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := Client{BaseURL: server.URL, TokenProvider: func(context.Context) (string, error) { return "runtime-token", nil }}
	page := (Page{Limit: 10, Cursor: "next"}).Values()
	resp, err := c.Do(context.Background(), http.MethodGet, "/v1/items?"+page.Encode(), io.Reader(nil), "request-1")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	got := NegotiateCapabilities(resp)
	if got.APIVersion != "1" || len(got.Features) != 2 {
		t.Fatalf("capabilities = %#v", got)
	}
}

func TestClientDoJSONProvidesTypedTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "invalid request headers", http.StatusBadRequest)
			return
		}
		var request map[string]string
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil || request["id"] != "env-1" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		_, _ = io.WriteString(w, `{"status":"ready"}`)
	}))
	defer server.Close()
	var response struct {
		Status string `json:"status"`
	}
	client := Client{BaseURL: server.URL}
	if err := client.DoJSON(context.Background(), http.MethodPost, "/v1/environments", map[string]string{"id": "env-1"}, &response, "request-1"); err != nil {
		t.Fatal(err)
	}
	if response.Status != "ready" {
		t.Fatalf("response = %#v", response)
	}
}

func TestClientKeepsOldMinorClientCompatibleWithAdditiveResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"ready","newMinorField":"ignored-by-old-client"}`)
	}))
	defer server.Close()
	var response struct {
		Status string `json:"status"`
	}
	if err := (Client{BaseURL: server.URL}).DoJSON(context.Background(), http.MethodGet, "/v1/items", nil, &response, ""); err != nil {
		t.Fatal(err)
	}
	if response.Status != "ready" {
		t.Fatalf("old client failed additive response compatibility: %#v", response)
	}
}

func TestClientReturnsStructuredAPIErrorWithoutRawPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"code":"quota_exhausted","feature":"environments","limit":"maxActiveEnvironments","current":2,"requested":1,"plan":"free","request_id":"req-1","secret":"must-not-be-retained"}`)
	}))
	defer server.Close()
	err := (Client{BaseURL: server.URL}).DoJSON(context.Background(), http.MethodPost, "/v1/items", map[string]string{}, nil, "idempotent-1")
	apiErr, ok := err.(*APIError)
	if !ok || apiErr.Status != http.StatusTooManyRequests || apiErr.Code != "quota_exhausted" || apiErr.RequestID != "req-1" {
		t.Fatalf("unexpected structured error: %#v", err)
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("API error exposed raw payload: %v", err)
	}
}

func TestVerifyWebhook(t *testing.T) {
	digest := hmac.New(sha256.New, []byte("secret"))
	_, _ = digest.Write([]byte("payload"))
	if !VerifyWebhook("secret", "sha256="+hex.EncodeToString(digest.Sum(nil)), []byte("payload")) {
		t.Fatal("valid signature rejected")
	}
	if VerifyWebhook("secret", "bad", []byte("payload")) {
		t.Fatal("invalid signature accepted")
	}
}
