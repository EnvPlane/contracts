package installationbilling

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"testing"
)

func TestProviderModeAudienceAndProofIsolation(t *testing.T) {
	c := loadCorpus(t)
	for _, p := range c.Positive {
		t.Run(p.Name, func(t *testing.T) {
			original, err := ParseVerifiedRequest([]byte(p.Request), 1800000000)
			if err != nil {
				t.Fatal(err)
			}
			live := original
			live.Intent.Scope.Mode, live.Intent.Scope.Audience = "live", LiveAudience
			if live.VerifyProof(1800000000) == nil {
				t.Fatal("test proof accepted in live mode")
			}
			var session *Session
			if p.Session != nil && ActionOperation(live.Intent.Operation) {
				v := *p.Session
				v.Scope = live.Intent.Scope
				session = &v
				f, e := session.Frame()
				if e != nil {
					t.Fatal(e)
				}
				live.Intent.SessionDigest = digest(f)
			}
			var next ed25519.PrivateKey
			if live.Intent.Operation == "key.rotate" {
				next = ed25519.NewKeyFromSeed(bytes.Repeat([]byte{42}, 32))
				live.Intent.NextPublicKey = b64(next.Public().(ed25519.PublicKey))
			}
			resign(&live)
			if next != nil {
				f, _ := live.ProofFrame()
				live.NextSignature = b64(ed25519.Sign(next, f))
			}
			raw, _ := json.Marshal(live)
			if _, err = ParseVerifiedRequest(raw, 1800000000); err != nil {
				t.Fatal("live proof", err)
			}
			if session != nil {
				if live.MatchSession(*session, 1800000000) != nil {
					t.Fatal("live descriptor mismatch")
				}
				if live.MatchSession(*p.Session, 1800000000) == nil || original.MatchSession(*session, 1800000000) == nil {
					t.Fatal("session crossed provider mode")
				}
			}
			a, _ := original.Intent.Frame()
			b, _ := live.Intent.Frame()
			if bytes.Equal(a, b) {
				t.Fatal("mode did not domain-separate intent")
			}
			live.Intent.Scope.Mode, live.Intent.Scope.Audience = "test", Audience
			if live.VerifyProof(1800000000) == nil {
				t.Fatal("live proof replayed in test mode")
			}
			for _, s := range []Scope{
				{Seller: original.Intent.Scope.Seller, Merchant: original.Intent.Scope.Merchant, Installation: original.Intent.Scope.Installation, Tenant: original.Intent.Scope.Tenant, Mode: "live", Audience: Audience},
				{Seller: original.Intent.Scope.Seller, Merchant: original.Intent.Scope.Merchant, Installation: original.Intent.Scope.Installation, Tenant: original.Intent.Scope.Tenant, Mode: "test", Audience: LiveAudience},
			} {
				if s.Validate() == nil {
					t.Fatal("foreign audience accepted")
				}
			}
		})
	}
}
