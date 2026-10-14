// Command installationpaidvectors emits synthetic public extension fixtures only.
package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	ib "github.com/envplane/contracts/installationbilling"
	"os"
	"strings"
)

func b64(b []byte) string  { return base64.RawURLEncoding.EncodeToString(b) }
func hash(b []byte) string { h := sha256.Sum256(b); return b64(h[:]) }
func js(v any) string      { b, _ := json.Marshal(v); return string(b) }
func main() {
	out := flag.String("out", "installationbilling/testdata/paid-v1.json", "output")
	check := flag.Bool("check", false, "check")
	flag.Parse()
	key := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{37}, 32))
	pub := key.Public().(ed25519.PublicKey)
	id, _ := ib.InstallationID(pub)
	scope := ib.Scope{Seller: "issuer-default", Merchant: "acct_fixture", Mode: "test", Audience: ib.Audience, Installation: id, Tenant: "tenant-a"}
	zero := b64(make([]byte, 32))
	handle := bytes.Repeat([]byte{41}, 32)
	positive := []map[string]any{}
	negative := []map[string]string{}
	for n, op := range []string{"session.paid.status", "session.redeem", "paid.status", "redeem"} {
		body := []byte("{}")
		if op == "paid.status" {
			body = []byte(js(ib.PaidStatusBody{Limit: 25}))
		}
		if op == "redeem" {
			body = []byte(js(ib.RedeemBody{AttemptID: "attempt-owned", Nonce: b64(bytes.Repeat([]byte{5}, 24))}))
		}
		i := ib.Intent{Scope: scope, Operation: op, Method: "POST", Target: ib.Routes()[op], BodyDigest: hash(body), CurrentEpoch: "1", CurrentPublicKey: b64(pub), NextEpoch: "0", NextPublicKey: zero, RequestID: b64(bytes.Repeat([]byte{byte(n + 40)}, 16)), SessionDigest: zero}
		session := ib.Session{Scope: scope, CurrentEpoch: "1", CurrentPublicKey: b64(pub), Operation: strings.TrimPrefix(op, "session."), HandleDigest: hash(handle), IssuedAt: "1800000000", ExpiresAt: "1800000300"}
		sf, e := session.Frame()
		if e != nil {
			panic(e)
		}
		if !strings.HasPrefix(op, "session.") {
			i.SessionDigest = hash(sf)
		}
		d, _ := i.Digest()
		ch := ib.Challenge{IntentDigest: d, Nonce: b64(bytes.Repeat([]byte{byte(n + 50)}, 32)), IssuedAt: "1800000000", ExpiresAt: "1800000300"}
		req := ib.Request{Version: ib.Version, Algorithm: ib.Algorithm, Intent: i, Body: b64(body), Challenge: ch}
		pf, _ := req.ProofFrame()
		req.Signature = b64(ed25519.Sign(key, pf))
		if !strings.HasPrefix(op, "session.") {
			req.SessionHandle = b64(handle)
		}
		if req.VerifyProof(1800000000) != nil {
			panic("invalid paid request")
		}
		resp := ib.Response{Version: ib.Version, Operation: op, RequestID: i.RequestID, IntentDigest: d, State: "ready", KeyEpoch: "1", Records: []ib.Record{}}
		if strings.HasPrefix(op, "session.") {
			resp.Session = &session
			resp.SessionHandle = b64(handle)
		} else {
			resp.State = "observed"
			resp.PaidStatus = &ib.PaidStatus{Records: []ib.PaidRecord{{AttemptID: "attempt-owned", PlanID: "team", PlanVersion: "2.0.0", SubscriptionState: "active", PaymentState: "paid", IssuanceState: "issued", State: "activation_required", PurchaseID: "stripe-test-sub_owned", LicenseID: "stripe-test-license", CurrentLicenseID: "stripe-test-license", ExpiresAt: "1801000000", GraceEndsAt: "0"}}}
			if op == "redeem" {
				resp.State = "redeemed"
				resp.ActivationCode = "v1.synthetic.fixture"
			}
		}
		if _, e = ib.ParseResponse([]byte(js(resp))); e != nil {
			panic(e)
		}
		fi, _ := i.Frame()
		positive = append(positive, map[string]any{"name": op, "request": js(req), "challengeRequest": js(ib.ChallengeRequest{Version: ib.Version, Algorithm: ib.Algorithm, Intent: i, Body: req.Body}), "challengeResponse": js(ib.ChallengeResponse{Version: ib.Version, Algorithm: ib.Algorithm, Challenge: ch}), "response": js(resp), "intentFrameHex": hex.EncodeToString(fi), "proofFrameHex": hex.EncodeToString(pf), "intentDigest": d, "session": session})
		bad := req
		bad.Intent.Scope.Mode = "live"
		negative = append(negative, map[string]string{"name": op + "-live", "phase": "request", "raw": js(bad)})
		bad = req
		bad.Intent.Target = ib.Routes()["checkout"]
		negative = append(negative, map[string]string{"name": op + "-route", "phase": "request", "raw": js(bad)})
		if resp.PaidStatus != nil {
			badResp := resp
			badResp.PaidStatus = &ib.PaidStatus{Records: []ib.PaidRecord{resp.PaidStatus.Records[0]}}
			badResp.PaidStatus.Records[0].PaymentState = "not_charged"
			negative = append(negative, map[string]string{"name": op + "-unpaid-proof", "phase": "response", "raw": js(badResp)})
		}
	}
	for _, v := range negative {
		var e error
		if v["phase"] == "request" {
			_, e = ib.ParseRequest([]byte(v["raw"]))
		} else {
			_, e = ib.ParseResponse([]byte(v["raw"]))
		}
		if e == nil {
			panic("negative accepted")
		}
	}
	b, _ := json.MarshalIndent(map[string]any{"version": ib.Version, "now": "1800000000", "positive": positive, "negative": negative}, "", "  ")
	b = append(b, '\n')
	if *check {
		old, _ := os.ReadFile(*out)
		if !bytes.Equal(old, b) {
			panic("paid vector drift")
		}
	} else if e := os.WriteFile(*out, b, 0600); e != nil {
		panic(e)
	}
	fmt.Printf("%d positive / %d negative paid vectors\n", len(positive), len(negative))
}
