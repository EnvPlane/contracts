package envplanesdk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientAuthPaginationAndIdempotency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer runtime-token" || r.Header.Get("Idempotency-Key") != "request-1" {
			t.Fatalf("headers not propagated")
		}
		if r.URL.Query().Get("cursor") != "next" || r.URL.Query().Get("limit") != "10" {
			t.Fatalf("pagination not encoded: %s", r.URL.RawQuery)
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
