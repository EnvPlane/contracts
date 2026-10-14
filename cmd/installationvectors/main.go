// Command installationvectors freezes public synthetic interoperability fixtures.
// Deterministic seeds below are public test data, never runtime credentials.
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"filippo.io/edwards25519"
	"flag"
	"fmt"
	ib "github.com/envplane/contracts/installationbilling"
	"os"
	"sort"
	"strings"
)

type positive struct {
	Name              string      `json:"name"`
	Request           string      `json:"request"`
	ChallengeRequest  string      `json:"challengeRequest"`
	ChallengeResponse string      `json:"challengeResponse"`
	Response          string      `json:"response"`
	IntentFrame       string      `json:"intentFrameHex"`
	ProofFrame        string      `json:"proofFrameHex"`
	IntentDigest      string      `json:"intentDigest"`
	Session           *ib.Session `json:"session"`
}
type negative struct {
	Name  string `json:"name"`
	Phase string `json:"phase"`
	Raw   string `json:"raw"`
}
type corpus struct {
	Version  string     `json:"version"`
	Now      string     `json:"now"`
	Positive []positive `json:"positive"`
	Negative []negative `json:"negative"`
}

func b64(b []byte) string  { return base64.RawURLEncoding.EncodeToString(b) }
func hash(b []byte) string { h := sha256.Sum256(b); return b64(h[:]) }
func js(v any) string {
	b, e := json.Marshal(v)
	if e != nil {
		panic(e)
	}
	return string(b)
}
func main() {
	out := flag.String("out", "installationbilling/testdata/v1.json", "output")
	check := flag.Bool("check", false, "check drift without writing")
	flag.Parse()
	c := corpus{Version: ib.Version, Now: "1800000000"}
	var keys []ed25519.PrivateKey
	for n := 0; n < 2; n++ {
		seed := make([]byte, 32)
		for i := range seed {
			seed[i] = byte(n*32 + i)
		}
		keys = append(keys, ed25519.NewKeyFromSeed(seed))
	}
	pub := keys[0].Public().(ed25519.PublicKey)
	id, e := ib.InstallationID(pub)
	if e != nil {
		panic(e)
	}
	scope := ib.Scope{Seller: "issuer-default", Merchant: "acct_fixture", Mode: "test", Audience: ib.Audience, Installation: id, Tenant: "tenant-a"}
	zero := b64(make([]byte, 32))
	handle := bytes.Repeat([]byte{71}, 32)
	ops := []string{}
	for op := range ib.Routes() {
		ops = append(ops, op)
	}
	sort.Strings(ops)
	var enroll ib.Request
	for n, op := range ops {
		raw := []byte("{}")
		switch op {
		case "checkout":
			raw = []byte(js(ib.CheckoutBody{AttemptID: "attempt-a", PlanID: "pro", PlanVersion: "2026.1", ReturnRoute: "billing"}))
		case "portal":
			raw = []byte(js(ib.PortalBody{ReturnRoute: "billing"}))
		case "status":
			raw = []byte(js(ib.StatusBody{AttemptID: "attempt-a", Cursor: "", Limit: 25}))
		case "key.status":
			raw = []byte(js(ib.KeyStatusBody{TransitionRequestID: b64(bytes.Repeat([]byte{1}, 16))}))
		}
		i := ib.Intent{Scope: scope, Operation: op, Method: "POST", Target: ib.Routes()[op], BodyDigest: hash(raw), CurrentEpoch: "1", CurrentPublicKey: b64(pub), NextEpoch: "0", NextPublicKey: zero, RequestID: b64(bytes.Repeat([]byte{byte(n + 1)}, 16)), SessionDigest: zero}
		if op == "key.rotate" {
			i.NextEpoch = "2"
			i.NextPublicKey = b64(keys[1].Public().(ed25519.PublicKey))
		}
		var session *ib.Session
		if op == "checkout" || op == "portal" || op == "status" || strings.HasPrefix(op, "session.") {
			s := ib.Session{Scope: scope, CurrentEpoch: "1", CurrentPublicKey: b64(pub), Operation: strings.TrimPrefix(op, "session."), HandleDigest: hash(handle), IssuedAt: "1800000000", ExpiresAt: "1800000300"}
			session = &s
			if !strings.HasPrefix(op, "session.") {
				f, e := s.Frame()
				if e != nil {
					panic(e)
				}
				i.SessionDigest = hash(f)
			}
		}
		d, e := i.Digest()
		if e != nil {
			panic(e)
		}
		ch := ib.Challenge{IntentDigest: d, Nonce: b64(bytes.Repeat([]byte{byte(n + 1)}, 32)), IssuedAt: "1800000000", ExpiresAt: "1800000300"}
		r := ib.Request{Version: ib.Version, Algorithm: ib.Algorithm, Intent: i, Body: b64(raw), Challenge: ch}
		f, e := r.ProofFrame()
		if e != nil {
			panic(e)
		}
		r.Signature = b64(ed25519.Sign(keys[0], f))
		if op == "key.rotate" {
			r.NextSignature = b64(ed25519.Sign(keys[1], f))
		}
		if op == "checkout" || op == "portal" || op == "status" {
			r.SessionHandle = b64(handle)
		}
		if r.VerifyProof(1800000000) != nil {
			panic("invalid positive")
		}
		if op == "enroll" {
			enroll = r
		}
		resp := ib.Response{Version: ib.Version, Operation: op, RequestID: i.RequestID, IntentDigest: d, KeyEpoch: "1", Records: []ib.Record{}}
		switch {
		case op == "enroll":
			resp.State = "registered"
		case op == "tenant.bind":
			resp.State = "bound"
		case strings.HasPrefix(op, "session."):
			resp.State = "ready"
			resp.Session = session
			resp.SessionHandle = b64(handle)
		case op == "checkout":
			resp.State = "ready"
			resp.URL = "https://checkout.stripe.com/c/pay/offline-fixture"
		case op == "portal":
			resp.State = "ready"
			resp.URL = "https://billing.stripe.com/p/session/offline-fixture"
		case op == "status":
			resp.State = "observed"
			resp.Records = []ib.Record{{AttemptID: "attempt-a", State: "checkout_pending"}}
		case op == "key.status":
			resp.State = "observed"
		case op == "key.rotate":
			resp.State = "rotated"
			resp.KeyEpoch = "2"
		case op == "key.revoke":
			resp.State = "revoked"
		}
		if _, e := ib.ParseResponse([]byte(js(resp))); e != nil {
			panic("invalid response " + op)
		}
		fi, _ := i.Frame()
		c.Positive = append(c.Positive, positive{Name: op, Request: js(r), ChallengeRequest: js(ib.ChallengeRequest{Version: ib.Version, Algorithm: ib.Algorithm, Intent: i, Body: r.Body}), ChallengeResponse: js(ib.ChallengeResponse{Version: ib.Version, Algorithm: ib.Algorithm, Challenge: ch}), Response: js(resp), IntentFrame: hex.EncodeToString(fi), ProofFrame: hex.EncodeToString(f), IntentDigest: d, Session: session})
	}
	add := func(name, phase, raw string) { c.Negative = append(c.Negative, negative{name, phase, raw}) }
	base := js(enroll)
	mut := func(name string, fn func(*ib.Request)) { r := enroll; fn(&r); add(name, "request", js(r)) }
	add("duplicate-top", "request", strings.Replace(base, `"version":`, `"version":"duplicate","version":`, 1))
	add("duplicate-escaped-key", "request", strings.Replace(base, `"version":`, `"\u0076ersion":"duplicate","version":`, 1))
	add("case-alias", "request", strings.Replace(base, `"version":`, `"Version":`, 1))
	add("unknown-top", "request", strings.Replace(base, `{"version":`, `{"unknown":1,"version":`, 1))
	add("missing-field", "request", strings.Replace(base, `"nextSignature":"",`, "", 1))
	add("trailing-json", "request", base+"{}")
	add("null-top", "request", "null")
	add("null-nested", "request", strings.Replace(base, `"scope":`+js(enroll.Intent.Scope), `"scope":null`, 1))
	add("lone-surrogate", "request", strings.Replace(base, `"tenant-a"`, `"\ud800"`, 1))
	mut("version", func(r *ib.Request) { r.Version = "v2" })
	mut("algorithm", func(r *ib.Request) { r.Algorithm = "Ed25519ph" })
	mut("live", func(r *ib.Request) { r.Intent.Scope.Mode = "live" })
	mut("foreign-audience", func(r *ib.Request) { r.Intent.Scope.Audience = "foreign" })
	mut("merchant-control", func(r *ib.Request) { r.Intent.Scope.Merchant = "acct_fixture\n" })
	mut("foreign-id", func(r *ib.Request) { r.Intent.Scope.Installation = "epi1_ed25519_" + strings.Repeat("0", 64) })
	mut("epoch-number-spelling", func(r *ib.Request) { r.Intent.CurrentEpoch = "01" })
	mut("epoch-overflow", func(r *ib.Request) { r.Intent.CurrentEpoch = "18446744073709551616" })
	mut("epoch-above-policy", func(r *ib.Request) { r.Intent.CurrentEpoch = "9223372036854775808" })
	mut("method", func(r *ib.Request) { r.Intent.Method = "GET" })
	mut("private-target", func(r *ib.Request) { r.Intent.Target = "/v1/internal/billing/checkout" })
	mut("query-target", func(r *ib.Request) { r.Intent.Target += "?x=1" })
	mut("padded-base64", func(r *ib.Request) { r.Signature += "=" })
	mut("short-signature", func(r *ib.Request) { r.Signature = b64(make([]byte, 63)) })
	mut("extra-rotation-proof", func(r *ib.Request) { r.NextSignature = r.Signature })
	mut("unexpected-handle", func(r *ib.Request) { r.SessionHandle = b64(handle) })
	mut("unknown-body-field", func(r *ib.Request) {
		r.Body = b64([]byte(`{"customerId":"cus_foreign"}`))
		r.Intent.BodyDigest = hash([]byte(`{"customerId":"cus_foreign"}`))
	})
	mut("small-order-key", func(r *ib.Request) { r.Intent.CurrentPublicKey = b64(append([]byte{1}, make([]byte, 31)...)) })
	mut("noncanonical-point", func(r *ib.Request) {
		b := bytes.Repeat([]byte{255}, 32)
		b[0] = 237
		b[31] = 127
		r.Intent.CurrentPublicKey = b64(b)
	})
	mut("invalid-curve-point", func(r *ib.Request) { b := make([]byte, 32); b[0] = 2; r.Intent.CurrentPublicKey = b64(b) })
	mut("bad-window", func(r *ib.Request) { r.Challenge.ExpiresAt = "1800000301" })
	mut("zero-nonce", func(r *ib.Request) { r.Challenge.Nonce = zero })
	mut("mixed-order-key", func(r *ib.Request) {
		a, _ := new(edwards25519.Point).SetBytes(pub)
		t := bytes.Repeat([]byte{255}, 32)
		t[0] = 236
		t[31] = 127
		b, _ := new(edwards25519.Point).SetBytes(t)
		r.Intent.CurrentPublicKey = b64(new(edwards25519.Point).Add(a, b).Bytes())
	})

	proofMut := func(name string, fn func(*ib.Request)) { r := enroll; fn(&r); add(name, "proof", js(r)) }
	proofMut("signature-tamper", func(r *ib.Request) {
		s, _ := base64.RawURLEncoding.DecodeString(r.Signature)
		s[0] ^= 1
		r.Signature = b64(s)
	})
	proofMut("signature-small-order-R", func(r *ib.Request) {
		s, _ := base64.RawURLEncoding.DecodeString(r.Signature)
		copy(s[:32], append([]byte{1}, make([]byte, 31)...))
		r.Signature = b64(s)
	})
	proofMut("noncanonical-scalar", func(r *ib.Request) {
		s, _ := base64.RawURLEncoding.DecodeString(r.Signature)
		for i := 32; i < 64; i++ {
			s[i] = 255
		}
		r.Signature = b64(s)
	})
	proofMut("future-window", func(r *ib.Request) { r.Challenge.IssuedAt = "1800000031"; r.Challenge.ExpiresAt = "1800000331" })
	proofMut("expired-window", func(r *ib.Request) { r.Challenge.IssuedAt = "1799999700"; r.Challenge.ExpiresAt = "1800000000" })
	for _, p := range c.Positive {
		if p.Name == "status" {
			add("fractional-limit", "request", strings.Replace(p.Request, `"body":"`+bodyOf(p.Request)+`"`, `"body":"`+b64([]byte(`{"attemptId":"attempt-a","cursor":"","limit":1e0}`))+`"`, 1))
		}
		if p.Name == "portal" {
			add("foreign-host", "response", strings.Replace(p.Response, "billing.stripe.com", "billing.stripe.com.evil", 1))
			add("unknown-response-field", "response", strings.Replace(p.Response, `{"version":`, `{"customerId":"cus_foreign","version":`, 1))
			add("response-null-records", "response", strings.Replace(p.Response, `"records":[]`, `"records":null`, 1))
		}
	}
	for _, x := range c.Negative {
		var err error
		switch x.Phase {
		case "request":
			_, err = ib.ParseRequest([]byte(x.Raw))
		case "proof":
			var r ib.Request
			r, err = ib.ParseRequest([]byte(x.Raw))
			if err == nil {
				err = r.VerifyProof(1800000000)
			}
		case "response":
			_, err = ib.ParseResponse([]byte(x.Raw))
		}
		if err == nil {
			panic("negative accepted " + x.Name)
		}
	}
	b, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		panic(e)
	}
	b = append(b, '\n')
	if *check {
		old, e := os.ReadFile(*out)
		if e != nil || !bytes.Equal(old, b) {
			panic("vector drift")
		}
	} else if e = os.WriteFile(*out, b, 0600); e != nil {
		panic(e)
	}
	fmt.Printf("%d positive, %d negative vectors\n", len(c.Positive), len(c.Negative))
}
func bodyOf(raw string) string {
	var r ib.Request
	if err := json.Unmarshal([]byte(raw), &r); err != nil {
		panic(err)
	}
	return r.Body
}
