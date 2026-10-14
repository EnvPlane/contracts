package installationbilling

import (
	"net/url"
	"strings"
)

// Response is a closed discriminated envelope. Unused values are explicit:
// session=null, sessionHandle="", url="", records=[], cursor="".
type Response struct {
	Version       string   `json:"version"`
	Operation     string   `json:"operation"`
	RequestID     string   `json:"requestId"`
	IntentDigest  string   `json:"intentDigest"`
	State         string   `json:"state"`
	KeyEpoch      string   `json:"keyEpoch"`
	Session       *Session `json:"session"`
	SessionHandle string   `json:"sessionHandle"`
	URL           string   `json:"url"`
	Records       []Record `json:"records"`
	Cursor        string   `json:"cursor"`
}
type Record struct {
	AttemptID  string `json:"attemptId"`
	State      string `json:"state"`
	PurchaseID string `json:"purchaseId"`
	LicenseID  string `json:"licenseId"`
}
type ErrorResponse struct {
	Version   string `json:"version"`
	Code      string `json:"code"`
	RequestID string `json:"requestId"`
}

func ErrorStatus(code string) int {
	switch code {
	case "invalid_request", "unsupported_version", "unsupported_algorithm", "invalid_proof":
		return 400
	case "scope_rejected", "authority_revoked", "key_epoch_stale", "session_rejected":
		return 403
	case "challenge_expired", "session_expired":
		return 410
	case "request_conflict", "challenge_consumed":
		return 409
	case "rate_limited":
		return 429
	case "temporarily_unavailable":
		return 503
	case "not_found":
		return 404
	default:
		return 0
	}
}
func ParseError(raw []byte) (ErrorResponse, error) {
	var e ErrorResponse
	if strict(raw, &e) != nil || e.Version != Version || ErrorStatus(e.Code) == 0 {
		return ErrorResponse{}, ErrInvalid
	}
	if e.RequestID != "" {
		if _, err := decode(e.RequestID, 16, true); err != nil {
			return ErrorResponse{}, ErrInvalid
		}
	}
	return e, nil
}
func ParseChallengeResponse(raw []byte) (ChallengeResponse, error) {
	var r ChallengeResponse
	if strict(raw, &r) != nil || r.Version != Version || r.Algorithm != Algorithm {
		return ChallengeResponse{}, ErrInvalid
	}
	if _, e := r.Challenge.Frame(); e != nil {
		return ChallengeResponse{}, ErrInvalid
	}
	return r, nil
}
func ParseResponse(raw []byte) (Response, error) {
	var r Response
	if strict(raw, &r) != nil || r.Version != Version || Routes()[r.Operation] == "" {
		return Response{}, ErrInvalid
	}
	if _, e := decode(r.RequestID, 16, true); e != nil {
		return Response{}, ErrInvalid
	}
	if _, e := decode(r.IntentDigest, 32, false); e != nil {
		return Response{}, ErrInvalid
	}
	if _, e := epoch(r.KeyEpoch); e != nil {
		return Response{}, ErrInvalid
	}
	allowed := map[string]string{"enroll": "registered", "tenant.bind": "bound", "session.checkout": "ready", "session.portal": "ready", "session.status": "ready", "portal": "ready", "status": "observed", "key.rotate": "rotated", "key.revoke": "revoked", "key.status": "observed"}
	if r.Operation == "checkout" {
		if r.State != "accepted" && r.State != "pending" && r.State != "ready" && r.State != "review_required" && r.State != "failed" {
			return Response{}, ErrInvalid
		}
	} else if r.State != allowed[r.Operation] {
		return Response{}, ErrInvalid
	}
	if strings.HasPrefix(r.Operation, "session.") {
		if r.Session == nil {
			return Response{}, ErrInvalid
		}
		if _, e := r.Session.Frame(); e != nil {
			return Response{}, ErrInvalid
		}
		if r.Session.Operation != strings.TrimPrefix(r.Operation, "session.") || r.Session.CurrentEpoch != r.KeyEpoch {
			return Response{}, ErrInvalid
		}
		if _, e := decode(r.SessionHandle, 32, true); e != nil || digest(bin(r.SessionHandle)) != r.Session.HandleDigest {
			return Response{}, ErrInvalid
		}
	} else if r.Session != nil || r.SessionHandle != "" {
		return Response{}, ErrInvalid
	}
	needsURL := r.Operation == "portal" || (r.Operation == "checkout" && r.State == "ready")
	if needsURL {
		u, e := url.Parse(r.URL)
		host := "checkout.stripe.com"
		if r.Operation == "portal" {
			host = "billing.stripe.com"
		}
		if e != nil || u.Scheme != "https" || u.Host != host || u.User != nil || u.Fragment != "" || u.Path == "" || !strings.HasPrefix(r.URL, "https://"+host+"/") || strings.Contains(r.URL, "#") || len(r.URL) > 4096 || strings.ContainsAny(r.URL, "\r\n\\") {
			return Response{}, ErrInvalid
		}
	} else if r.URL != "" {
		return Response{}, ErrInvalid
	}
	if r.Records == nil || len(r.Records) > 25 || len(r.Cursor) > 1024 || (r.Cursor != "" && !validText(r.Cursor, 1024)) {
		return Response{}, ErrInvalid
	}
	if r.Operation != "status" && (len(r.Records) != 0 || r.Cursor != "") {
		return Response{}, ErrInvalid
	}
	seen := map[string]bool{}
	for _, x := range r.Records {
		if !validText(x.AttemptID, 128) || seen[x.AttemptID] || (x.PurchaseID != "" && !validText(x.PurchaseID, 128)) || (x.LicenseID != "" && !validText(x.LicenseID, 128)) {
			return Response{}, ErrInvalid
		}
		seen[x.AttemptID] = true
		switch x.State {
		case "checkout_pending", "payment_processing", "paid/issuance_pending", "activation_required", "past_due", "canceled", "failed":
		default:
			return Response{}, ErrInvalid
		}
	}
	return r, nil
}

// MatchResponse prevents applying a delayed/foreign response to another request.
func (r Request) MatchResponse(v Response) error {
	d, e := r.Intent.Digest()
	if e != nil || v.Operation != r.Intent.Operation || v.RequestID != r.Intent.RequestID || v.IntentDigest != d {
		return ErrInvalid
	}
	expected := r.Intent.CurrentEpoch
	if v.Operation == "key.rotate" {
		expected = r.Intent.NextEpoch
	}
	if v.KeyEpoch != expected {
		return ErrInvalid
	}
	if v.Session != nil && (v.Session.Scope != r.Intent.Scope || v.Session.CurrentPublicKey != r.Intent.CurrentPublicKey) {
		return ErrInvalid
	}
	return nil
}
