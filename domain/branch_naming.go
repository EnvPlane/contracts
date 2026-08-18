package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"unicode"
)

const maxDNSLabelLength = 63

// BranchEnvironmentName is the canonical cross-language naming result. ID is
// immutable after creation; DisplayName may change when a branch is renamed.
type BranchEnvironmentName struct {
	ID            string   `json:"id"`
	DisplayName   string   `json:"displayName"`
	Namespace     string   `json:"namespace"`
	ReleaseName   string   `json:"releaseName"`
	IngressName   string   `json:"ingressName"`
	Compatibility []string `json:"compatibilityAliases,omitempty"`
}

func BranchEnvironmentNameFor(project, repository, branch, provider, changeID string) BranchEnvironmentName {
	display := BranchDisplayName(branch)
	identity := strings.Join([]string{strings.ToLower(strings.TrimSpace(project)), strings.ToLower(strings.TrimSpace(provider)), strings.ToLower(strings.TrimSpace(repository)), strings.TrimSpace(changeID)}, "|")
	sum := sha256.Sum256([]byte(identity))
	hash := hex.EncodeToString(sum[:])[:8]
	providerID := NormalizeEnvironmentID(provider)
	if providerID == "" {
		providerID = "scm"
	}
	prefix := strings.Join([]string{display, providerID, hash}, "-")
	id := BoundedDNSName(prefix, maxDNSLabelLength)
	if id == "" {
		id = BoundedDNSName(providerID+"-"+hash, maxDNSLabelLength)
	}
	aliases := []string{}
	if changeID != "" {
		legacyPrefix := "pr-"
		if strings.EqualFold(provider, string(ProviderGitLab)) {
			legacyPrefix = "mr-"
		}
		legacy := NormalizeEnvironmentID(legacyPrefix + changeID)
		if legacy != "" && legacy != id {
			aliases = append(aliases, legacy)
		}
	}
	return BranchEnvironmentName{ID: id, DisplayName: display, Namespace: id, ReleaseName: id, IngressName: id, Compatibility: aliases}
}

func BranchDisplayName(branch string) string {
	value := strings.ToLower(strings.TrimSpace(branch))
	value = strings.NewReplacer("ä", "a", "ö", "o", "ü", "u", "ß", "ss", "é", "e", "è", "e", "à", "a").Replace(value)
	var out strings.Builder
	lastDash := false
	for _, r := range value {
		if r > unicode.MaxASCII {
			continue
		}
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			out.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			out.WriteByte('-')
			lastDash = true
		}
	}
	display := strings.Trim(out.String(), "-")
	if display == "" || display == "default" || display == "system" || display == "kube-system" {
		display = "branch"
	}
	return BoundedDNSName(display, 40)
}

func BoundedDNSName(value string, max int) string {
	value = NormalizeEnvironmentID(value)
	if max <= 0 || max > maxDNSLabelLength {
		max = maxDNSLabelLength
	}
	if len(value) <= max {
		return value
	}
	sum := sha256.Sum256([]byte(value))
	suffix := hex.EncodeToString(sum[:])[:8]
	cut := max - len(suffix) - 1
	if cut < 1 {
		return suffix[:max]
	}
	return strings.TrimRight(value[:cut], "-") + "-" + suffix
}

func (n BranchEnvironmentName) Valid() bool {
	for _, value := range []string{n.ID, n.Namespace, n.ReleaseName, n.IngressName} {
		if value == "" || len(value) > maxDNSLabelLength || NormalizeEnvironmentID(value) != value {
			return false
		}
	}
	return true
}
