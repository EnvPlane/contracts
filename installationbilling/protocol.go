// Package installationbilling defines the opt-in public installation billing v1
// contract. Validation is not admission: callers must match durable challenges,
// registered keys/epochs, session authority and financial/runtime guards.
package installationbilling

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"filippo.io/edwards25519"
)

const Version = "envplane.install-billing.v1"
const Algorithm = "Ed25519"
const Audience = "envplane.install-billing.test.v1"
const MaxEnvelope = 65536
const MaxBody = 32768
const TTL uint64 = 300
const ClockSkew uint64 = 30
const MaxEpoch uint64 = 1<<63 - 1
const ChallengePath = "/v1/installation-billing/challenge"

var ErrInvalid = errors.New("installation billing contract rejected")

// Routes returns a fresh operation map; changing it cannot alter validation.
func Routes() map[string]string {
	return map[string]string{
		"enroll":           "/v1/installation-billing/enroll",
		"session.checkout": "/v1/installation-billing/sessions/checkout",
		"session.portal":   "/v1/installation-billing/sessions/portal",
		"session.status":   "/v1/installation-billing/sessions/status",
		"checkout":         "/v1/installation-billing/checkout",
		"portal":           "/v1/installation-billing/portal",
		"status":           "/v1/installation-billing/status",
		"tenant.bind":      "/v1/installation-billing/tenants/bind",
		"key.rotate":       "/v1/installation-billing/keys/rotate",
		"key.revoke":       "/v1/installation-billing/keys/revoke",
		"key.status":       "/v1/installation-billing/keys/status",
	}
}

type Scope struct {
	Seller       string `json:"seller"`
	Merchant     string `json:"merchant"`
	Mode         string `json:"mode"`
	Audience     string `json:"audience"`
	Installation string `json:"installation"`
	Tenant       string `json:"tenant"`
}
type Intent struct {
	Scope            Scope  `json:"scope"`
	Operation        string `json:"operation"`
	Method           string `json:"method"`
	Target           string `json:"target"`
	BodyDigest       string `json:"bodyDigest"`
	CurrentEpoch     string `json:"currentEpoch"`
	CurrentPublicKey string `json:"currentPublicKey"`
	NextEpoch        string `json:"nextEpoch"`
	NextPublicKey    string `json:"nextPublicKey"`
	RequestID        string `json:"requestId"`
	SessionDigest    string `json:"sessionDigest"`
}
type Challenge struct {
	IntentDigest string `json:"intentDigest"`
	Nonce        string `json:"nonce"`
	IssuedAt     string `json:"issuedAt"`
	ExpiresAt    string `json:"expiresAt"`
}
type ChallengeRequest struct {
	Version   string `json:"version"`
	Algorithm string `json:"algorithm"`
	Intent    Intent `json:"intent"`
	Body      string `json:"body"`
}
type ChallengeResponse struct {
	Version   string    `json:"version"`
	Algorithm string    `json:"algorithm"`
	Challenge Challenge `json:"challenge"`
}
type Request struct {
	Version       string    `json:"version"`
	Algorithm     string    `json:"algorithm"`
	Intent        Intent    `json:"intent"`
	Body          string    `json:"body"`
	Challenge     Challenge `json:"challenge"`
	Signature     string    `json:"signature"`
	NextSignature string    `json:"nextSignature"`
	SessionHandle string    `json:"sessionHandle"`
}
type Session struct {
	Scope            Scope  `json:"scope"`
	CurrentEpoch     string `json:"currentEpoch"`
	CurrentPublicKey string `json:"currentPublicKey"`
	Operation        string `json:"operation"`
	HandleDigest     string `json:"handleDigest"`
	IssuedAt         string `json:"issuedAt"`
	ExpiresAt        string `json:"expiresAt"`
}
type CheckoutBody struct {
	AttemptID   string `json:"attemptId"`
	PlanID      string `json:"planId"`
	PlanVersion string `json:"planVersion"`
	ReturnRoute string `json:"returnRoute"`
}
type PortalBody struct {
	ReturnRoute string `json:"returnRoute"`
}
type StatusBody struct {
	AttemptID string `json:"attemptId"`
	Cursor    string `json:"cursor"`
	Limit     int    `json:"limit"`
}
type KeyStatusBody struct {
	TransitionRequestID string `json:"transitionRequestId"`
}

var accountRE = regexp.MustCompile(`^acct_[A-Za-z0-9_]{1,250}$`)
var installationRE = regexp.MustCompile(`^epi1_ed25519_[0-9a-f]{64}$`)

func validText(s string, max int) bool {
	if s == "" || len(s) > max || !utf8.ValidString(s) || strings.TrimSpace(s) != s {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}
func (s Scope) Validate() error {
	if s.Seller != "issuer-default" || !accountRE.MatchString(s.Merchant) || s.Mode != "test" || s.Audience != Audience || !installationRE.MatchString(s.Installation) || !validText(s.Tenant, 128) {
		return ErrInvalid
	}
	return nil
}
func decimal(s string, zero bool) (uint64, error) {
	v, e := strconv.ParseUint(s, 10, 64)
	if e != nil || strconv.FormatUint(v, 10) != s || (!zero && v == 0) {
		return 0, ErrInvalid
	}
	return v, nil
}
func epoch(s string) (uint64, error) {
	v, e := decimal(s, false)
	if e != nil || v > MaxEpoch {
		return 0, ErrInvalid
	}
	return v, nil
}
func decode(s string, n int, nonzero bool) ([]byte, error) {
	b, e := base64.RawURLEncoding.Strict().DecodeString(s)
	if e != nil || len(b) != n || base64.RawURLEncoding.EncodeToString(b) != s || (nonzero && bytes.Equal(b, make([]byte, n))) {
		return nil, ErrInvalid
	}
	return b, nil
}
func digest(raw []byte) string {
	h := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(h[:])
}
func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
func point(raw []byte) bool {
	p, e := new(edwards25519.Point).SetBytes(raw)
	if e != nil || !bytes.Equal(p.Bytes(), raw) {
		return false
	}
	cleared := new(edwards25519.Point).MultByCofactor(p)
	if cleared.Equal(edwards25519.NewIdentityPoint()) == 1 {
		return false
	}
	eight := make([]byte, 32)
	eight[0] = 8
	scalar, _ := edwards25519.NewScalar().SetCanonicalBytes(eight)
	inverse := edwards25519.NewScalar().Invert(scalar)
	return new(edwards25519.Point).ScalarMult(inverse, cleared).Equal(p) == 1
}
func validSignature(pub, msg, sig []byte) bool {
	if len(sig) != 64 || !point(pub) || !point(sig[:32]) {
		return false
	}
	if _, e := edwards25519.NewScalar().SetCanonicalBytes(sig[32:]); e != nil {
		return false
	}
	return ed25519.Verify(pub, msg, sig)
}
func frame(domain string, fields ...[]byte) []byte {
	if len(domain) > 65535 || len(fields) > 65535 {
		panic("installation frame bounds")
	}
	b := bytes.NewBufferString("ENVPLANE-INSTALL-BILL\x00\x01")
	if err := binary.Write(b, binary.BigEndian, uint16(len(domain))); err != nil { // #nosec G115 -- explicit uint16 bound checked above.
		panic("installation frame encoding")
	}
	b.WriteString(domain)
	if err := binary.Write(b, binary.BigEndian, uint16(len(fields))); err != nil { // #nosec G115 -- explicit uint16 bound checked above.
		panic("installation frame encoding")
	}
	for _, f := range fields {
		if uint64(len(f)) > 4294967295 {
			panic("installation frame bounds")
		}
		if err := binary.Write(b, binary.BigEndian, uint32(len(f))); err != nil { // #nosec G115 -- explicit uint32 bound checked above.
			panic("installation frame encoding")
		}
		b.Write(f)
	}
	return b.Bytes()
}
func num(s string) []byte {
	v, _ := decimal(s, true)
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
func bin(s string) []byte { b, _ := base64.RawURLEncoding.DecodeString(s); return b }
func (s Scope) fields() [][]byte {
	return [][]byte{[]byte(s.Seller), []byte(s.Merchant), []byte(s.Mode), []byte(s.Audience), []byte(s.Installation), []byte(s.Tenant)}
}
func InstallationID(pub []byte) (string, error) {
	if !point(pub) {
		return "", ErrInvalid
	}
	h := sha256.Sum256(frame("genesis", pub))
	return "epi1_ed25519_" + hex.EncodeToString(h[:]), nil
}
func (i Intent) Validate() error {
	if i.Scope.Validate() != nil || Routes()[i.Operation] == "" || i.Target != Routes()[i.Operation] || i.Method != "POST" {
		return ErrInvalid
	}
	current, e := epoch(i.CurrentEpoch)
	if e != nil {
		return ErrInvalid
	}
	pub, e := decode(i.CurrentPublicKey, 32, true)
	if e != nil || !point(pub) {
		return ErrInvalid
	}
	for _, s := range []string{i.BodyDigest, i.SessionDigest} {
		if _, e = decode(s, 32, false); e != nil {
			return ErrInvalid
		}
	}
	if _, e = decode(i.RequestID, 16, true); e != nil {
		return ErrInvalid
	}
	next, e := decimal(i.NextEpoch, true)
	if e != nil {
		return ErrInvalid
	}
	newKey, e := decode(i.NextPublicKey, 32, false)
	if e != nil {
		return ErrInvalid
	}
	if i.Operation == "key.rotate" {
		if current == MaxEpoch || next != current+1 || !point(newKey) || bytes.Equal(pub, newKey) {
			return ErrInvalid
		}
	} else if next != 0 || !bytes.Equal(newKey, make([]byte, 32)) {
		return ErrInvalid
	}
	action := i.Operation == "checkout" || i.Operation == "portal" || i.Operation == "status"
	session, _ := decode(i.SessionDigest, 32, false)
	if action == bytes.Equal(session, make([]byte, 32)) {
		return ErrInvalid
	}
	if i.Operation == "enroll" {
		id, _ := InstallationID(pub)
		if current != 1 || id != i.Scope.Installation {
			return ErrInvalid
		}
	}
	return nil
}
func (i Intent) Frame() ([]byte, error) {
	if i.Validate() != nil {
		return nil, ErrInvalid
	}
	f := i.Scope.fields()
	f = append(f, []byte(i.Operation), []byte(i.Method), []byte(i.Target), bin(i.BodyDigest), num(i.CurrentEpoch), bin(i.CurrentPublicKey), num(i.NextEpoch), bin(i.NextPublicKey), bin(i.RequestID), bin(i.SessionDigest))
	return frame("intent", f...), nil
}
func (i Intent) Digest() (string, error) {
	b, e := i.Frame()
	if e != nil {
		return "", e
	}
	return digest(b), nil
}
func validWindow(issued, expires string) bool {
	a, e := decimal(issued, false)
	if e != nil {
		return false
	}
	b, e := decimal(expires, false)
	return e == nil && a <= ^uint64(0)-TTL && b == a+TTL
}
func windowNow(issued, expires string, now uint64) bool {
	a, _ := decimal(issued, false)
	b, _ := decimal(expires, false)
	return validWindow(issued, expires) && now < b && (a <= now || a-now <= ClockSkew)
}
func (c Challenge) Frame() ([]byte, error) {
	if _, e := decode(c.IntentDigest, 32, false); e != nil {
		return nil, ErrInvalid
	}
	if _, e := decode(c.Nonce, 32, true); e != nil || !validWindow(c.IssuedAt, c.ExpiresAt) {
		return nil, ErrInvalid
	}
	return frame("challenge", bin(c.IntentDigest), bin(c.Nonce), num(c.IssuedAt), num(c.ExpiresAt)), nil
}
func proofDomain(op string) string {
	switch {
	case op == "enroll":
		return "enroll-proof"
	case strings.HasPrefix(op, "session."):
		return "session-proof"
	case op == "key.rotate":
		return "rotate-proof"
	case op == "key.revoke":
		return "revoke-proof"
	default:
		return "action-proof"
	}
}
func (r Request) ProofFrame() ([]byte, error) {
	id, e := r.Intent.Digest()
	if e != nil || id != r.Challenge.IntentDigest {
		return nil, ErrInvalid
	}
	ch, e := r.Challenge.Frame()
	if e != nil {
		return nil, ErrInvalid
	}
	return frame(proofDomain(r.Intent.Operation), bin(id), bin(digest(ch))), nil
}
func (s Session) Frame() ([]byte, error) {
	if s.Scope.Validate() != nil {
		return nil, ErrInvalid
	}
	if _, e := epoch(s.CurrentEpoch); e != nil {
		return nil, ErrInvalid
	}
	pub, e := decode(s.CurrentPublicKey, 32, true)
	if e != nil || !point(pub) {
		return nil, ErrInvalid
	}
	if s.Operation != "checkout" && s.Operation != "portal" && s.Operation != "status" {
		return nil, ErrInvalid
	}
	if _, e = decode(s.HandleDigest, 32, true); e != nil || !validWindow(s.IssuedAt, s.ExpiresAt) {
		return nil, ErrInvalid
	}
	return frame("session", append(s.Scope.fields(), num(s.CurrentEpoch), pub, []byte(s.Operation), bin(s.HandleDigest), num(s.IssuedAt), num(s.ExpiresAt))...), nil
}

// MatchSession checks descriptor/key binding only; storage must also verify active
// authority and atomically admit work. Expiry here cannot invalidate admitted work.
func (r Request) MatchSession(s Session, now uint64) error {
	if _, e := decode(r.SessionHandle, 32, true); e != nil || r.Intent.Validate() != nil {
		return ErrInvalid
	}
	f, e := s.Frame()
	if e != nil || !windowNow(s.IssuedAt, s.ExpiresAt, now) || s.Scope != r.Intent.Scope || s.Operation != r.Intent.Operation || s.CurrentEpoch != r.Intent.CurrentEpoch || s.CurrentPublicKey != r.Intent.CurrentPublicKey || digest(f) != r.Intent.SessionDigest || digest(bin(r.SessionHandle)) != s.HandleDigest {
		return ErrInvalid
	}
	return nil
}
func body(raw string) ([]byte, error) {
	b, e := base64.RawURLEncoding.Strict().DecodeString(raw)
	if e != nil || len(b) > MaxBody || b64(b) != raw {
		return nil, ErrInvalid
	}
	return b, nil
}
func ValidateBody(op string, raw []byte) error {
	if len(raw) > MaxBody {
		return ErrInvalid
	}
	switch op {
	case "checkout":
		var b CheckoutBody
		if strict(raw, &b) != nil || !validText(b.AttemptID, 128) || !validText(b.PlanID, 128) || !validText(b.PlanVersion, 128) || b.ReturnRoute != "billing" {
			return ErrInvalid
		}
	case "portal":
		var b PortalBody
		if strict(raw, &b) != nil || b.ReturnRoute != "billing" {
			return ErrInvalid
		}
	case "status":
		var b StatusBody
		if strict(raw, &b) != nil || !validText(b.AttemptID, 128) || len(b.Cursor) > 1024 || (b.Cursor != "" && !validText(b.Cursor, 1024)) || b.Limit < 1 || b.Limit > 25 {
			return ErrInvalid
		}
	case "key.status":
		var b KeyStatusBody
		if strict(raw, &b) != nil {
			return ErrInvalid
		}
		if _, e := decode(b.TransitionRequestID, 16, true); e != nil {
			return ErrInvalid
		}
	default:
		if Routes()[op] == "" {
			return ErrInvalid
		}
		var empty struct{}
		if strict(raw, &empty) != nil {
			return ErrInvalid
		}
	}
	return nil
}
func common(version, algorithm string, i Intent, raw string) error {
	if version != Version || algorithm != Algorithm || i.Validate() != nil {
		return ErrInvalid
	}
	b, e := body(raw)
	if e != nil || digest(b) != i.BodyDigest {
		return ErrInvalid
	}
	return ValidateBody(i.Operation, b)
}
func ParseChallengeRequest(raw []byte) (ChallengeRequest, error) {
	var r ChallengeRequest
	if strict(raw, &r) != nil || common(r.Version, r.Algorithm, r.Intent, r.Body) != nil {
		return ChallengeRequest{}, ErrInvalid
	}
	return r, nil
}
func ParseRequest(raw []byte) (Request, error) {
	var r Request
	if strict(raw, &r) != nil || common(r.Version, r.Algorithm, r.Intent, r.Body) != nil {
		return Request{}, ErrInvalid
	}
	if _, e := r.ProofFrame(); e != nil {
		return Request{}, ErrInvalid
	}
	if _, e := decode(r.Signature, 64, false); e != nil {
		return Request{}, ErrInvalid
	}
	if r.Intent.Operation == "key.rotate" {
		if _, e := decode(r.NextSignature, 64, false); e != nil {
			return Request{}, ErrInvalid
		}
	} else if r.NextSignature != "" {
		return Request{}, ErrInvalid
	}
	action := r.Intent.Operation == "checkout" || r.Intent.Operation == "portal" || r.Intent.Operation == "status"
	if action {
		if _, e := decode(r.SessionHandle, 32, true); e != nil {
			return Request{}, ErrInvalid
		}
	} else if r.SessionHandle != "" {
		return Request{}, ErrInvalid
	}
	return r, nil
}

// VerifyProof verifies stateless contract integrity only. Never use it as an
// enrollment/session decision without matching trusted persisted authority.
func (r Request) VerifyProof(now uint64) error {
	raw, e := json.Marshal(r)
	if e != nil {
		return ErrInvalid
	}
	r, e = ParseRequest(raw)
	if e != nil || !windowNow(r.Challenge.IssuedAt, r.Challenge.ExpiresAt, now) {
		return ErrInvalid
	}
	f, _ := r.ProofFrame()
	if !validSignature(bin(r.Intent.CurrentPublicKey), f, bin(r.Signature)) {
		return ErrInvalid
	}
	if r.Intent.Operation == "key.rotate" && !validSignature(bin(r.Intent.NextPublicKey), f, bin(r.NextSignature)) {
		return ErrInvalid
	}
	return nil
}
func ValidateHTTP(method, target, encoding string) error {
	if method != "POST" || (encoding != "" && encoding != "identity") {
		return ErrInvalid
	}
	if target == ChallengePath {
		return nil
	}
	for _, p := range Routes() {
		if target == p {
			return nil
		}
	}
	return ErrInvalid
}

// strict rejects duplicate/case-alias/unknown/missing/null fields before Go's
// otherwise case-insensitive struct decoding, with bounded bytes and depth.
func strict(raw []byte, out any) error {
	if len(raw) == 0 || len(raw) > MaxEnvelope || !utf8.Valid(raw) || !validUnicodeJSON(raw) {
		return ErrInvalid
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	node, e := readJSON(d, 0)
	if e != nil || d.Decode(new(any)) != io.EOF || shape(node, reflect.TypeOf(out).Elem()) != nil {
		return ErrInvalid
	}
	d = json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(out) != nil {
		return ErrInvalid
	}
	return nil
}
func shape(v any, t reflect.Type) error {
	if t.Kind() == reflect.Pointer {
		if v == nil {
			return nil
		}
		return shape(v, t.Elem())
	}
	if v == nil {
		return ErrInvalid
	}
	if t.Kind() == reflect.Slice {
		a, ok := v.([]any)
		if !ok {
			return ErrInvalid
		}
		for _, x := range a {
			if shape(x, t.Elem()) != nil {
				return ErrInvalid
			}
		}
		return nil
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	obj, ok := v.(map[string]any)
	if !ok || len(obj) != t.NumField() {
		return ErrInvalid
	}
	for n := 0; n < t.NumField(); n++ {
		f := t.Field(n)
		key := strings.Split(f.Tag.Get("json"), ",")[0]
		value, ok := obj[key]
		if !ok || shape(value, f.Type) != nil {
			return ErrInvalid
		}
	}
	return nil
}
func readJSON(d *json.Decoder, depth int) (any, error) {
	if depth > 12 {
		return nil, ErrInvalid
	}
	tok, e := d.Token()
	if e != nil {
		return nil, ErrInvalid
	}
	switch x := tok.(type) {
	case json.Delim:
		switch x {
		case '{':
			m := map[string]any{}
			for d.More() {
				k, e := d.Token()
				key, ok := k.(string)
				if e != nil || !ok {
					return nil, ErrInvalid
				}
				if _, ok = m[key]; ok {
					return nil, ErrInvalid
				}
				v, e := readJSON(d, depth+1)
				if e != nil {
					return nil, e
				}
				m[key] = v
			}
			end, e := d.Token()
			if e != nil || end != json.Delim('}') {
				return nil, ErrInvalid
			}
			return m, nil
		case '[':
			a := []any{}
			for d.More() {
				v, e := readJSON(d, depth+1)
				if e != nil {
					return nil, e
				}
				a = append(a, v)
			}
			end, e := d.Token()
			if e != nil || end != json.Delim(']') {
				return nil, ErrInvalid
			}
			return a, nil
		default:
			return nil, ErrInvalid
		}
	case json.Number:
		if strings.ContainsAny(string(x), ".eE") {
			return nil, ErrInvalid
		}
		return tok, nil
	case string:
		if !utf8.ValidString(x) {
			return nil, ErrInvalid
		}
		return x, nil
	default:
		return tok, nil
	}
}

func validUnicodeJSON(raw []byte) bool {
	for i := 0; i < len(raw); i++ {
		if raw[i] != '\\' {
			continue
		}
		i++
		if i >= len(raw) {
			return false
		}
		if raw[i] != 'u' {
			continue
		}
		if i+4 >= len(raw) {
			return false
		}
		v, e := strconv.ParseUint(string(raw[i+1:i+5]), 16, 16)
		if e != nil {
			return false
		}
		i += 4
		if v >= 0xdc00 && v <= 0xdfff {
			return false
		}
		if v >= 0xd800 && v <= 0xdbff {
			if i+6 >= len(raw) || raw[i+1] != '\\' || raw[i+2] != 'u' {
				return false
			}
			w, e := strconv.ParseUint(string(raw[i+3:i+7]), 16, 16)
			if e != nil || w < 0xdc00 || w > 0xdfff {
				return false
			}
			i += 6
		}
	}
	return true
}

func ParseVerifiedRequest(raw []byte, now uint64) (Request, error) {
	r, e := ParseRequest(raw)
	if e != nil || r.VerifyProof(now) != nil {
		return Request{}, ErrInvalid
	}
	return r, nil
}
