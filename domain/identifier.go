package domain

import (
	"regexp"
	"strings"
)

// EnvironmentWorkBranch returns the canonical GitOps work branch for an
// environment. The prefix is deliberately module-independent and stable
// across control-plane, gitops, and runner implementations.
func EnvironmentWorkBranch(environmentID string) string {
	return "envplane/env-" + NormalizeEnvironmentID(environmentID)
}

var invalidEnvironmentIDCharacters = regexp.MustCompile(`[^a-z0-9-]+`)

// NormalizeEnvironmentID converts an arbitrary external identifier to a DNS-safe label.
func NormalizeEnvironmentID(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = invalidEnvironmentIDCharacters.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if len(value) > 63 {
		value = strings.TrimRight(value[:63], "-")
	}
	return value
}
