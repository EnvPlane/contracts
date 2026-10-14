package installationbilling

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type wireCorpus struct {
	Version  string `json:"version"`
	Now      string `json:"now"`
	Positive []struct {
		Name              string   `json:"name"`
		Request           string   `json:"request"`
		ChallengeRequest  string   `json:"challengeRequest"`
		ChallengeResponse string   `json:"challengeResponse"`
		Response          string   `json:"response"`
		IntentFrame       string   `json:"intentFrameHex"`
		ProofFrame        string   `json:"proofFrameHex"`
		IntentDigest      string   `json:"intentDigest"`
		Session           *Session `json:"session"`
	} `json:"positive"`
	Negative []struct {
		Name  string `json:"name"`
		Phase string `json:"phase"`
		Raw   string `json:"raw"`
	} `json:"negative"`
}

func loadCorpus(t *testing.T) wireCorpus {
	t.Helper()
	raw, e := os.ReadFile("testdata/v1.json")
	if e != nil {
		t.Fatal(e)
	}
	var c wireCorpus
	if e = json.Unmarshal(raw, &c); e != nil {
		t.Fatal(e)
	}
	if c.Version != Version || len(c.Positive) != 11 || len(c.Negative) < 40 {
		t.Fatal("incomplete corpus")
	}
	return c
}
func TestWireVectors(t *testing.T) {
	c := loadCorpus(t)
	for _, p := range c.Positive {
		t.Run(p.Name, func(t *testing.T) {
			r, e := ParseRequest([]byte(p.Request))
			if e != nil || r.VerifyProof(1800000000) != nil {
				t.Fatal("request", e)
			}
			f, _ := r.Intent.Frame()
			if hex.EncodeToString(f) != p.IntentFrame {
				t.Fatal("intent frame drift")
			}
			f, _ = r.ProofFrame()
			if hex.EncodeToString(f) != p.ProofFrame {
				t.Fatal("proof frame drift")
			}
			d, _ := r.Intent.Digest()
			if d != p.IntentDigest {
				t.Fatal("digest drift")
			}
			if _, e = ParseChallengeRequest([]byte(p.ChallengeRequest)); e != nil {
				t.Fatal(e)
			}
			if _, e = ParseChallengeResponse([]byte(p.ChallengeResponse)); e != nil {
				t.Fatal(e)
			}
			v, e := ParseResponse([]byte(p.Response))
			if e != nil || r.MatchResponse(v) != nil {
				t.Fatal("response binding", e)
			}
			v.RequestID = b64(bytes.Repeat([]byte{255}, 16))
			if r.MatchResponse(v) == nil {
				t.Fatal("foreign response")
			}
			if _, e = ParseVerifiedRequest([]byte(p.Request), 1800000000); e != nil {
				t.Fatal(e)
			}
			if r.SessionHandle != "" && r.MatchSession(*p.Session, 1800000000) != nil {
				t.Fatal("session binding")
			}
			if ValidateHTTP("POST", r.Intent.Target, "identity") != nil {
				t.Fatal("route")
			}
		})
	}
	for _, n := range c.Negative {
		t.Run(n.Name, func(t *testing.T) {
			var e error
			switch n.Phase {
			case "request":
				_, e = ParseRequest([]byte(n.Raw))
			case "proof":
				var r Request
				r, e = ParseRequest([]byte(n.Raw))
				if e == nil {
					e = r.VerifyProof(1800000000)
				}
			case "response":
				_, e = ParseResponse([]byte(n.Raw))
			default:
				t.Fatal("unknown phase")
			}
			if e == nil {
				t.Fatal("negative accepted")
			}
		})
	}
}
func resign(r *Request) {
	d, _ := r.Intent.Digest()
	r.Challenge.IntentDigest = d
	f, _ := r.ProofFrame()
	seed := make([]byte, 32)
	for i := range seed {
		seed[i] = byte(i)
	}
	r.Signature = b64(ed25519.Sign(ed25519.NewKeyFromSeed(seed), f))
}
func TestTimeBodyScopeAndRotationFailures(t *testing.T) {
	c := loadCorpus(t)
	for _, p := range c.Positive {
		r, _ := ParseRequest([]byte(p.Request))
		if r.VerifyProof(1800000300) == nil || r.VerifyProof(1799999969) == nil {
			t.Fatal("time boundary")
		}
		if r.VerifyProof(1799999970) != nil {
			t.Fatal("skew boundary")
		}
		if r.SessionHandle != "" {
			s := *p.Session
			s.Scope.Tenant = "foreign"
			if r.MatchSession(s, 1800000000) == nil {
				t.Fatal("foreign session")
			}
			s = *p.Session
			s.CurrentEpoch = "2"
			if r.MatchSession(s, 1800000000) == nil {
				t.Fatal("old epoch")
			}
			if r.MatchSession(*p.Session, 1800000300) == nil {
				t.Fatal("expired session")
			}
		}
	}
	r, _ := ParseRequest([]byte(c.Positive[1].Request))
	r.Body = b64([]byte(`{"customerId":"cus_foreign"}`))
	r.Intent.BodyDigest = digest(bin(r.Body))
	resign(&r)
	raw, _ := json.Marshal(r)
	if _, e := ParseRequest(raw); e == nil {
		t.Fatal("unknown body")
	}
	for _, p := range c.Positive {
		if p.Name != "key.rotate" {
			continue
		}
		r, _ := ParseRequest([]byte(p.Request))
		r.NextSignature = r.Signature
		if r.VerifyProof(1800000000) == nil {
			t.Fatal("rotation needs new key proof")
		}
		r.Intent.CurrentEpoch = "9223372036854775807"
		r.Intent.NextEpoch = "9223372036854775808"
		if r.Intent.Validate() == nil {
			t.Fatal("epoch overflow")
		}
	}
}
func TestStrictJSONBoundsAndHostileKeys(t *testing.T) {
	c := loadCorpus(t)
	base := c.Positive[0].Request
	for _, raw := range [][]byte{bytes.Repeat([]byte{' '}, MaxEnvelope+1), append([]byte(base), 255), []byte(strings.Replace(base, `"tenant-a"`, `"\udc00"`, 1)), []byte(strings.Replace(base, `"tenant-a"`, `"\ud800x"`, 1))} {
		if _, e := ParseRequest(raw); e == nil {
			t.Fatal("bad JSON accepted")
		}
	}
	for _, body := range []string{`{"attemptId":"a","cursor":"","limit":1.0}`, `{"attemptId":"a","cursor":"","limit":1e0}`, `{"attemptId":"a","cursor":"","limit":null}`, `{"attemptId":"a","cursor":"","limit":1,"Limit":1}`} {
		if ValidateBody("status", []byte(body)) == nil {
			t.Fatal("numeric/alias ambiguity")
		}
	}
	for _, p := range [][]byte{make([]byte, 32), append([]byte{1}, make([]byte, 31)...), bytes.Repeat([]byte{255}, 32), append([]byte{2}, make([]byte, 31)...)} {
		if _, e := InstallationID(p); e == nil {
			t.Fatal("hostile key accepted", hex.EncodeToString(p))
		}
	}
	for _, target := range []string{ChallengePath + "?x=1", "/v1/internal/billing/checkout", ChallengePath + "/", "/v1/installation-billing/%63hallenge"} {
		if ValidateHTTP("POST", target, "") == nil {
			t.Fatal("route accepted")
		}
	}
	if ValidateHTTP("GET", ChallengePath, "") == nil || ValidateHTTP("POST", ChallengePath, "gzip") == nil {
		t.Fatal("HTTP mismatch")
	}
	for _, code := range []string{"invalid_request", "temporarily_unavailable", "challenge_consumed"} {
		raw, _ := json.Marshal(ErrorResponse{Version, code, ""})
		if _, e := ParseError(raw); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := ParseError([]byte(`{"version":"envplane.install-billing.v1","code":"raw_provider_error","requestId":""}`)); e == nil {
		t.Fatal("unbounded error")
	}
}

func TestDecimalUint64FramesHaveNoPrecisionLoss(t *testing.T) {
	c := Challenge{IntentDigest: b64(make([]byte, 32)), Nonce: b64(bytes.Repeat([]byte{1}, 32)), IssuedAt: "18446744073709551315", ExpiresAt: "18446744073709551615"}
	f, e := c.Frame()
	if e != nil || !bytes.Equal(f[len(f)-8:], bytes.Repeat([]byte{255}, 8)) {
		t.Fatal("uint64 lost precision", e)
	}
	for _, bad := range []string{"1e3", "1.0", "01", "1\n", "+1", "18446744073709551616"} {
		if _, e := decimal(bad, true); e == nil {
			t.Fatal("decimal accepted", bad)
		}
	}
	r, _ := ParseRequest([]byte(loadCorpus(t).Positive[0].Request))
	r.Intent.CurrentEpoch = "9007199254740993"
	f, e = r.Intent.Frame()
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Contains(f, num("9007199254740993")) {
		t.Fatal("epoch rounded")
	}
}
