package installationbilling

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPaidExtensionWireVectors(t *testing.T) {
	raw, e := os.ReadFile("testdata/paid-v1.json")
	if e != nil {
		t.Fatal(e)
	}
	var c wireCorpus
	if json.Unmarshal(raw, &c) != nil || len(c.Positive) != 4 || len(c.Negative) < 10 {
		t.Fatal("incomplete extension vectors")
	}
	for _, p := range c.Positive {
		t.Run(p.Name, func(t *testing.T) {
			r, e := ParseVerifiedRequest([]byte(p.Request), 1800000000)
			if e != nil {
				t.Fatal(e)
			}
			f, _ := r.Intent.Frame()
			if hex.EncodeToString(f) != p.IntentFrame {
				t.Fatal("frame drift")
			}
			f, _ = r.ProofFrame()
			if hex.EncodeToString(f) != p.ProofFrame {
				t.Fatal("proof drift")
			}
			v, e := ParseResponse([]byte(p.Response))
			if e != nil || r.MatchResponse(v) != nil {
				t.Fatal("response mismatch", e)
			}
			if r.SessionHandle != "" && r.MatchSession(*p.Session, 1800000000) != nil {
				t.Fatal("session mismatch")
			}
		})
	}
	for _, n := range c.Negative {
		var e error
		if n.Phase == "request" {
			_, e = ParseRequest([]byte(n.Raw))
		} else {
			_, e = ParseResponse([]byte(n.Raw))
		}
		if e == nil {
			t.Fatal("negative accepted", n.Name)
		}
	}
}
func TestPaidExtensionRejectsCallerLicenseUnknownBodyAndUnpaidGrant(t *testing.T) {
	for _, op := range []string{"paid.status", "redeem"} {
		for _, raw := range []string{`{"attemptId":"owned","licenseId":"foreign","nonce":"bad"}`, `{"attemptId":"owned","attemptId":"foreign"}`, `null`, `[]`} {
			if ValidateBody(op, []byte(raw)) == nil {
				t.Fatal("ambiguous body", op)
			}
		}
	}
	c := loadCorpus(t)
	for _, v := range c.Positive {
		var r Response
		if json.Unmarshal([]byte(v.Response), &r) != nil {
			t.Fatal("fixture")
		}
		bad := strings.Replace(v.Response, `"records":`, `"paidStatus":null,"records":`, 1)
		if _, e := ParseResponse([]byte(bad)); e == nil {
			t.Fatal("old response accepted paid extension field")
		}
	}
}
