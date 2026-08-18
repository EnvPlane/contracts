package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// RenderEnvironmentReleasePlan is the single deterministic transformation
// path for all delivery backends. It operates on sanitized manifests only.
func RenderEnvironmentReleasePlan(revision EnvironmentTemplateRevision, graph ServiceGraph, inputs EnvironmentRenderInputs) (EnvironmentReleasePlan, error) {
	if err := revision.Validate(); err != nil {
		return EnvironmentReleasePlan{}, err
	}
	if graph.Validation == nil || !graph.Validation.Valid {
		return EnvironmentReleasePlan{}, fmt.Errorf("dependency graph is unresolved")
	}
	if strings.TrimSpace(inputs.TenantID) == "" || inputs.TenantID != revision.TenantID || inputs.ProjectID != revision.ProjectID || strings.TrimSpace(inputs.EnvironmentID) == "" {
		return EnvironmentReleasePlan{}, fmt.Errorf("render input tenant/project binding is invalid")
	}
	if strings.TrimSpace(inputs.CommitSHA) == "" {
		return EnvironmentReleasePlan{}, fmt.Errorf("render input commit sha is required")
	}
	nodes := map[string]ServiceGraphNode{}
	for _, node := range graph.Nodes {
		nodes[node.ID] = node
	}
	policies := map[string]ResourceDependencyPolicy{}
	for _, policy := range graph.Policies {
		policies[policy.ResourceID] = policy
	}
	for _, issue := range graph.Validation.Errors {
		return EnvironmentReleasePlan{}, fmt.Errorf("dependency graph unresolved: %s", issue.Message)
	}
	for _, edge := range graph.Edges {
		if edge.Required && !nodes[edge.To] {
			return EnvironmentReleasePlan{}, fmt.Errorf("unresolved required link %s at %s", edge.To, edge.Path)
		}
		if target, ok := nodes[edge.To]; ok && sourceNamespaceIsBase(target.Namespace, revision.SourceNamespaces) {
			if policy, ok := policies[edge.To]; ok && (policy.Strategy == ResourcePolicyReference || policy.Strategy == ResourcePolicyExternal) {
				return EnvironmentReleasePlan{}, fmt.Errorf("base link %s at %s requires clone or parameterize", edge.To, edge.Path)
			}
		}
	}
	resources := make([]RenderedResource, 0, len(revision.Resources))
	transformations := make([]TransformationReport, 0)
	outputs := make([]RenderOutput, 0)
	ownership := make([]OwnershipRecord, 0)
	endpoints := make([]IngressEndpoint, 0)
	nameMap := map[string]string{}
	for _, resource := range revision.Resources {
		id := resourceID(resource.Kind, resource.Namespace, resource.Name)
		name := inputs.ResourceNames[id]
		if name == "" {
			name = inputs.ResourceNames[resource.Name]
		}
		if name == "" {
			name = renderedName(inputs.EnvironmentID, resource.Name)
		}
		nameMap[id] = name
	}
	for _, resource := range revision.Resources {
		id := resourceID(resource.Kind, resource.Namespace, resource.Name)
		policy, ok := policies[id]
		if !ok {
			return EnvironmentReleasePlan{}, fmt.Errorf("missing resource policy for %s", id)
		}
		if policy.Strategy == ResourcePolicyUnsupported || policy.Strategy == ResourcePolicyReference || policy.Strategy == ResourcePolicyExternal {
			return EnvironmentReleasePlan{}, fmt.Errorf("resource %s is not independently renderable: strategy=%s", id, policy.Strategy)
		}
		if resource.Kind == "Secret" || resource.Kind == "PersistentVolumeClaim" {
			return EnvironmentReleasePlan{}, fmt.Errorf("resource %s materialization is deferred", id)
		}
		cloned := cloneRenderValue(resource.Manifest)
		manifest, ok := cloned.(map[string]any)
		if !ok {
			return EnvironmentReleasePlan{}, fmt.Errorf("resource %s manifest is not an object", id)
		}
		transformManifest(manifest, resource, revision, inputs, nodes, &transformations)
		rewriteNamedReferences(manifest, resource, inputs, nodes, nameMap, &transformations)
		if resource.Kind == "Ingress" {
			endpoint, err := RewriteIngressManifest(manifest, resource.Name, IngressRenderInput{EnvironmentID: inputs.EnvironmentID, Project: inputs.ProjectID, Branch: inputs.BranchSlug, ChangeID: inputs.ChangeID, DomainRoot: inputs.ProjectDomain, Policy: PreviewDomainPolicy{Pattern: inputs.PreviewDomainPattern}})
			if err != nil {
				return EnvironmentReleasePlan{}, fmt.Errorf("ingress %s: %w", id, err)
			}
			endpoints = append(endpoints, endpoint)
		}
		if containsBaseLink(manifest, revision.SourceNamespaces, inputs.NamespaceMap) {
			return EnvironmentReleasePlan{}, fmt.Errorf("unresolved base namespace link in %s", id)
		}
		if err := validateRenderedManifest(manifest, resource.Kind); err != nil {
			return EnvironmentReleasePlan{}, fmt.Errorf("%s: %w", id, err)
		}
		name := metadataString(manifest, "name")
		namespace := metadataString(manifest, "namespace")
		digest, _ := digestJSON(manifest)
		resources = append(resources, RenderedResource{ResourceID: id, Kind: resource.Kind, Name: name, Namespace: namespace, Manifest: manifest, Digest: digest})
		outputs = append(outputs, RenderOutput{Kind: resource.Kind, Name: name, Namespace: namespace, Digest: digest})
		ownership = append(ownership, OwnershipRecord{Kind: resource.Kind, Name: name, Namespace: namespace})
	}
	sort.Slice(resources, func(i, j int) bool { return resources[i].ResourceID < resources[j].ResourceID })
	sort.Slice(outputs, func(i, j int) bool {
		return outputs[i].Kind+outputs[i].Namespace+outputs[i].Name < outputs[j].Kind+outputs[j].Namespace+outputs[j].Name
	})
	sort.Slice(ownership, func(i, j int) bool {
		return ownership[i].Kind+ownership[i].Namespace+ownership[i].Name < ownership[j].Kind+ownership[j].Namespace+ownership[j].Name
	})
	sort.Slice(transformations, func(i, j int) bool {
		a := transformations[i].ResourceID + transformations[i].Path + transformations[i].Type
		b := transformations[j].ResourceID + transformations[j].Path + transformations[j].Type
		return a < b
	})
	inputDigest, err := digestJSON(inputs)
	if err != nil {
		return EnvironmentReleasePlan{}, err
	}
	plan := EnvironmentReleasePlan{ContractVersion: EnvironmentTemplateContractVersion, PlanID: revision.RevisionID + "/" + inputs.EnvironmentID, TenantID: inputs.TenantID, ProjectID: inputs.ProjectID, EnvironmentID: inputs.EnvironmentID, TemplateRevisionID: revision.RevisionID, TemplateDigest: revision.Digest, Backend: inputs.Backend, RenderedResources: resources, Transformations: transformations, Outputs: outputs, Ownership: ownership, InputDigest: inputDigest}
	endpoints = SortIngressEndpoints(endpoints)
	plan.Endpoints = endpoints
	if len(endpoints) > 0 {
		plan.PrimaryEndpoint = endpoints[0].Host
	}
	plan.Digest, err = plan.CanonicalDigest()
	if err != nil {
		return EnvironmentReleasePlan{}, err
	}
	return plan, nil
}

func resourceID(kind, namespace, name string) string {
	return strings.TrimSpace(kind) + "/" + strings.TrimSpace(namespace) + "/" + strings.TrimSpace(name)
}
func renderedName(environmentID, name string) string {
	value := strings.ToLower(strings.TrimSpace(environmentID) + "-" + strings.TrimSpace(name))
	var out strings.Builder
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			out.WriteRune(char)
		} else {
			out.WriteRune('-')
		}
	}
	return strings.Trim(out.String(), "-")
}
func cloneRenderValue(value any) any {
	switch item := value.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, v := range item {
			out[k] = cloneRenderValue(v)
		}
		return out
	case []any:
		out := make([]any, len(item))
		for i, v := range item {
			out[i] = cloneRenderValue(v)
		}
		return out
	default:
		return item
	}
}
func digestJSON(value any) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
func sourceNamespaceIsBase(namespace string, sourceNamespaces []string) bool {
	for _, source := range sourceNamespaces {
		if strings.TrimSpace(namespace) == strings.TrimSpace(source) {
			return true
		}
	}
	return false
}
func containsBaseLink(value any, sourceNamespaces []string, namespaceMap map[string]string) bool {
	encoded, _ := json.Marshal(value)
	text := string(encoded)
	for _, source := range sourceNamespaces {
		target := namespaceMap[source]
		if target == "" || target == source {
			if strings.Contains(text, "\"namespace\":\""+source+"\"") || strings.Contains(text, "."+source+".svc") {
				return true
			}
		}
	}
	return false
}
func metadataString(manifest map[string]any, key string) string {
	meta, _ := manifest["metadata"].(map[string]any)
	value, _ := meta[key].(string)
	return value
}
func validateRenderedManifest(manifest map[string]any, kind string) error {
	if strings.TrimSpace(fmt.Sprint(manifest["apiVersion"])) == "" || strings.TrimSpace(fmt.Sprint(manifest["kind"])) == "" {
		return fmt.Errorf("apiVersion and kind are required")
	}
	if kind != "Namespace" && metadataString(manifest, "namespace") == "" {
		return fmt.Errorf("namespace is required")
	}
	if metadataString(manifest, "name") == "" {
		return fmt.Errorf("metadata.name is required")
	}
	return nil
}
func transformManifest(manifest map[string]any, resource ResourceTemplate, revision EnvironmentTemplateRevision, inputs EnvironmentRenderInputs, nodes map[string]ServiceGraphNode, report *[]TransformationReport) {
	id := resourceID(resource.Kind, resource.Namespace, resource.Name)
	rewriteMap(manifest, "", id, resource, revision, inputs, nodes, report)
}
func rewriteMap(value map[string]any, path, id string, resource ResourceTemplate, revision EnvironmentTemplateRevision, inputs EnvironmentRenderInputs, nodes map[string]ServiceGraphNode, report *[]TransformationReport) {
	for key, raw := range value {
		current := path + "." + key
		switch child := raw.(type) {
		case map[string]any:
			rewriteMap(child, current, id, resource, revision, inputs, nodes, report)
		case []any:
			for i, item := range child {
				if nested, ok := item.(map[string]any); ok {
					rewriteMap(nested, fmt.Sprintf("%s[%d]", current, i), id, resource, revision, inputs, nodes, report)
				} else if text, ok := item.(string); ok {
					child[i] = transformString(text, current, id, resource, inputs, report)
				}
			}
		case string:
			value[key] = transformString(child, current, id, resource, inputs, report)
		}
	}
}
func transformString(text, path, id string, resource ResourceTemplate, inputs EnvironmentRenderInputs, report *[]TransformationReport) string {
	original := text
	result := text
	if path == ".metadata.namespace" || strings.HasSuffix(path, ".namespace") {
		if mapped := inputs.NamespaceMap[resource.Namespace]; mapped != "" {
			result = mapped
		}
	}
	if path == ".metadata.name" {
		if mapped := inputs.ResourceNames[id]; mapped != "" {
			result = mapped
		}
	}
	for source, target := range inputs.NamespaceMap {
		if source != "" && target != "" {
			result = strings.ReplaceAll(result, "."+source+".svc", "."+target+".svc")
			result = strings.ReplaceAll(result, "."+source+".svc.cluster.local", "."+target+".svc.cluster.local")
		}
	}
	for source, target := range inputs.Hostnames {
		result = strings.ReplaceAll(result, source, target)
	}
	if strings.Contains(path, ".host") && inputs.ProjectDomain != "" && result == original {
		result = inputs.EnvironmentID + "-" + resource.Name + "." + strings.TrimSpace(inputs.ProjectDomain)
	}
	if strings.Contains(path, ".image") {
		repository := result
		if at := strings.Index(repository, "@"); at >= 0 {
			repository = repository[:at]
		}
		if colon := strings.LastIndex(repository, ":"); colon > strings.LastIndex(repository, "/") {
			repository = repository[:colon]
		}
		if image := inputs.ComponentImages[id]; image != "" {
			result = image
		} else if image := inputs.ComponentImages[resource.Name]; image != "" {
			result = image
		} else {
			result = repository + ":" + inputs.CommitSHA
		}
		if inputs.ImmutableImages && strings.HasSuffix(result, ":latest") {
			result = repository + ":" + inputs.CommitSHA
		}
	}
	if result != original {
		from, _ := digestJSON(original)
		to, _ := digestJSON(result)
		*report = append(*report, TransformationReport{ResourceID: id, Path: path, Type: transformationType(path), FromDigest: from, ToDigest: to, Reason: "canonical feature render"})
	}
	return result
}
func transformationType(path string) string {
	lower := strings.ToLower(path)
	switch {
	case strings.Contains(lower, "image"):
		return "image"
	case strings.Contains(lower, "host") || strings.Contains(lower, "url") || strings.Contains(lower, "dsn"):
		return "dns-url"
	case strings.Contains(lower, "namespace"):
		return "namespace"
	case strings.Contains(lower, "name"):
		return "name"
	default:
		return "reference"
	}
}

func rewriteNamedReferences(manifest map[string]any, resource ResourceTemplate, inputs EnvironmentRenderInputs, nodes map[string]ServiceGraphNode, names map[string]string, report *[]TransformationReport) {
	var walk func(any, string)
	walk = func(value any, path string) {
		switch item := value.(type) {
		case map[string]any:
			for key, raw := range item {
				current := path + "." + key
				if text, ok := raw.(string); ok {
					result := text
					if key == "name" && path == ".metadata" {
						result = names[resourceID(resource.Kind, resource.Namespace, resource.Name)]
					}
					if key == "name" || key == "secretName" || key == "claimName" || key == "serviceAccountName" {
						for nodeID, node := range nodes {
							if node.Namespace == resource.Namespace && node.Name == text {
								if mapped := names[nodeID]; mapped != "" {
									result = mapped
									break
								}
							}
						}
					}
					for nodeID, node := range nodes {
						if mapped := names[nodeID]; mapped != "" {
							targetNamespace := inputs.NamespaceMap[node.Namespace]
							if targetNamespace != "" {
								result = strings.ReplaceAll(result, node.Name+"."+node.Namespace+".svc", mapped+"."+targetNamespace+".svc")
								result = strings.ReplaceAll(result, node.Name+"."+node.Namespace+".svc.cluster.local", mapped+"."+targetNamespace+".svc.cluster.local")
							}
						}
					}
					if result != text {
						from, _ := digestJSON(text)
						to, _ := digestJSON(result)
						*report = append(*report, TransformationReport{ResourceID: resourceID(resource.Kind, resource.Namespace, resource.Name), Path: current, Type: "name", FromDigest: from, ToDigest: to, Reason: "dependency identity rewrite"})
						item[key] = result
					}
				} else {
					walk(raw, current)
				}
			}
		case []any:
			for index, child := range item {
				walk(child, fmt.Sprintf("%s[%d]", path, index))
			}
		}
	}
	walk(manifest, "")
}
