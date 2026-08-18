package domain

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var previewPlaceholderPattern = regexp.MustCompile(`\{(project|branch|changeId|ingress|service|domain)\}`)
var safeIngressAnnotationPattern = regexp.MustCompile(`^(nginx\.ingress\.kubernetes\.io/(backend-protocol|proxy-body-size|proxy-connect-timeout|proxy-read-timeout|proxy-send-timeout|ssl-redirect)|cert-manager\.io/(cluster-issuer|issuer)|external-dns\.alpha\.kubernetes\.io/ttl)$`)
var unsafeIngressAnnotationPattern = regexp.MustCompile(`(?i)(snippet|configuration|server-alias|permanent-redirect|temporal-redirect|auth-url|mirror)`)

type PreviewDomainPolicy struct {
	Pattern              string   `json:"pattern"`
	WildcardDNS          bool     `json:"wildcardDns,omitempty"`
	WildcardCertificates bool     `json:"wildcardCertificates,omitempty"`
	AllowedAnnotations   []string `json:"allowedAnnotations,omitempty"`
}

type IngressEndpoint struct {
	SourceName       string            `json:"sourceName"`
	Name             string            `json:"name"`
	Host             string            `json:"host"`
	TLSSecretName    string            `json:"tlsSecretName,omitempty"`
	Paths            []string          `json:"paths,omitempty"`
	BackendServices  []string          `json:"backendServices,omitempty"`
	IngressClassName string            `json:"ingressClassName,omitempty"`
	Annotations      map[string]string `json:"annotations,omitempty"`
	Primary          bool              `json:"primary"`
}

type EndpointPreflight struct {
	Ready                bool     `json:"ready"`
	DomainPatternValid   bool     `json:"domainPatternValid"`
	WildcardDNSAvailable bool     `json:"wildcardDnsAvailable"`
	WildcardTLSAvailable bool     `json:"wildcardTlsAvailable"`
	MissingCapabilities  []string `json:"missingCapabilities,omitempty"`
	Reason               string   `json:"reason,omitempty"`
}

type IngressRenderInput struct {
	EnvironmentID string
	Project       string
	Branch        string
	ChangeID      string
	DomainRoot    string
	Policy        PreviewDomainPolicy
	Allowlist     []string
}

func RewriteIngressManifest(manifest map[string]any, sourceName string, input IngressRenderInput) (IngressEndpoint, error) {
	if input.Policy.Pattern == "" {
		input.Policy.Pattern = "{branch}-{ingress}.{domain}"
	}
	metadata, _ := manifest["metadata"].(map[string]any)
	if metadata == nil {
		return IngressEndpoint{}, fmt.Errorf("ingress metadata is required")
	}
	generatedName := BoundedDNSName(input.EnvironmentID+"-"+sourceName, 63)
	metadata["name"] = generatedName
	annotations, _ := metadata["annotations"].(map[string]any)
	metadata["annotations"] = SafeIngressAnnotations(annotations, input.Allowlist)
	spec, _ := manifest["spec"].(map[string]any)
	if spec == nil {
		return IngressEndpoint{}, fmt.Errorf("ingress spec is required")
	}
	endpoint := IngressEndpoint{SourceName: sourceName, Name: generatedName, Annotations: map[string]string{}}
	if class, ok := spec["ingressClassName"].(string); ok {
		endpoint.IngressClassName = class
	}
	rules, _ := spec["rules"].([]any)
	for index, raw := range rules {
		rule, _ := raw.(map[string]any)
		if rule == nil {
			continue
		}
		service, paths := "", []string{}
		if http, ok := rule["http"].(map[string]any); ok {
			if entries, ok := http["paths"].([]any); ok {
				for _, rawPath := range entries {
					if path, ok := rawPath.(map[string]any); ok {
						if value, ok := path["path"].(string); ok {
							paths = append(paths, value)
						}
						if backend, ok := path["backend"].(map[string]any); ok {
							if svc, ok := backend["service"].(map[string]any); ok {
								if name, ok := svc["name"].(string); ok {
									service = name
									endpoint.BackendServices = append(endpoint.BackendServices, name)
								}
							}
						}
					}
				}
			}
		}
		host, err := RenderPreviewHost(input.Policy, input.Project, input.Branch, input.ChangeID, sourceName, service, input.DomainRoot, input.EnvironmentID)
		if err != nil {
			return IngressEndpoint{}, err
		}
		rule["host"] = host
		if index == 0 {
			endpoint.Host = host
		}
		endpoint.Paths = append(endpoint.Paths, paths...)
	}
	if endpoint.Host == "" {
		return IngressEndpoint{}, fmt.Errorf("ingress %q has no rules", sourceName)
	}
	tls, _ := spec["tls"].([]any)
	secretName := BoundedDNSName(input.EnvironmentID+"-"+sourceName+"-tls", 63)
	for _, raw := range tls {
		if entry, ok := raw.(map[string]any); ok {
			entry["secretName"] = secretName
			hosts, _ := entry["hosts"].([]any)
			for i := range hosts {
				hosts[i] = endpoint.Host
			}
			entry["hosts"] = hosts
		}
	}
	endpoint.TLSSecretName = secretName
	return endpoint, nil
}

func ValidatePreviewDomainPolicy(policy PreviewDomainPolicy) error {
	pattern := strings.TrimSpace(policy.Pattern)
	if pattern == "" {
		return fmt.Errorf("preview domain pattern is required")
	}
	if !previewPlaceholderPattern.MatchString(pattern) || (!strings.Contains(pattern, "{branch}") && !strings.Contains(pattern, "{changeId}")) {
		return fmt.Errorf("preview domain pattern must include {branch} or {changeId}")
	}
	if strings.ContainsAny(pattern, " /\\\t\r\n") {
		return fmt.Errorf("preview domain pattern contains invalid characters")
	}
	for _, part := range strings.Split(pattern, ".") {
		if part == "" || strings.Contains(part, "*") {
			return fmt.Errorf("preview domain pattern contains an invalid DNS label")
		}
	}
	return nil
}

func RenderPreviewHost(policy PreviewDomainPolicy, project, branch, changeID, ingress, service, domainRoot, environmentID string) (string, error) {
	if err := ValidatePreviewDomainPolicy(policy); err != nil {
		return "", err
	}
	values := map[string]string{
		"project": NormalizeEnvironmentID(project), "branch": BranchDisplayName(branch),
		"changeId": NormalizeEnvironmentID(changeID), "ingress": NormalizeEnvironmentID(ingress),
		"service": NormalizeEnvironmentID(service), "domain": strings.Trim(strings.ToLower(domainRoot), "."),
	}
	host := policy.Pattern
	for key, value := range values {
		host = strings.ReplaceAll(host, "{"+key+"}", value)
	}
	host = strings.Trim(strings.ToLower(host), ".")
	if strings.Contains(host, "{") || strings.Contains(host, "}") {
		return "", fmt.Errorf("preview domain pattern has unresolved placeholder")
	}
	labels := strings.Split(host, ".")
	for i, label := range labels {
		labels[i] = BoundedDNSName(label, 63)
	}
	host = strings.Join(labels, ".")
	if netHost, err := url.Parse("https://" + host); err != nil || netHost.Host != host || host == "" {
		return "", fmt.Errorf("rendered preview host is invalid")
	}
	_ = environmentID
	return host, nil
}

func SafeIngressAnnotations(input map[string]any, allowlist []string) map[string]string {
	allowed := map[string]bool{}
	for _, key := range allowlist {
		allowed[strings.TrimSpace(key)] = true
	}
	out := map[string]string{}
	for key, value := range input {
		if unsafeIngressAnnotationPattern.MatchString(key) || (!safeIngressAnnotationPattern.MatchString(key) && !allowed[key]) {
			continue
		}
		out[key] = strings.TrimSpace(fmt.Sprint(value))
	}
	return out
}

func SortIngressEndpoints(endpoints []IngressEndpoint) []IngressEndpoint {
	sort.SliceStable(endpoints, func(i, j int) bool { return endpoints[i].SourceName < endpoints[j].SourceName })
	for i := range endpoints {
		endpoints[i].Primary = i == 0
	}
	return endpoints
}

func FeatureOwnedIngressArtifacts(environmentID string, endpoint IngressEndpoint) []OwnershipRecord {
	return []OwnershipRecord{{Kind: "Ingress", Name: endpoint.Name, Namespace: environmentID}, {Kind: "Secret", Name: endpoint.TLSSecretName, Namespace: environmentID}}
}

func IsFeatureOwnedIngressArtifact(environmentID, name string, endpoint IngressEndpoint) bool {
	for _, item := range FeatureOwnedIngressArtifacts(environmentID, endpoint) {
		if item.Name == name && item.Namespace == environmentID {
			return true
		}
	}
	return false
}
