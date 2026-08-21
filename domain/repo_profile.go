package domain

import (
	"errors"
	"sort"
	"strings"
)

const RepoProfileSchemaVersion = "1"

type RepoManifest struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

type RepoProfile struct {
	SchemaVersion       string         `json:"schemaVersion"`
	TenantID            string         `json:"tenantId"`
	ProjectID           string         `json:"projectId"`
	Manifests           []RepoManifest `json:"manifests,omitempty"`
	DeclaredPorts       []int          `json:"declaredPorts,omitempty"`
	EnvironmentNames    []string       `json:"environmentNames,omitempty"`
	ComponentCatalogIDs []string       `json:"componentCatalogIds,omitempty"`
	BootstrapFieldNames []string       `json:"bootstrapFieldNames,omitempty"`
}

func (p RepoProfile) Deterministic() RepoProfile {
	p.Manifests = append([]RepoManifest(nil), p.Manifests...)
	sort.Slice(p.Manifests, func(i, j int) bool {
		if p.Manifests[i].Path != p.Manifests[j].Path {
			return p.Manifests[i].Path < p.Manifests[j].Path
		}
		return p.Manifests[i].Type < p.Manifests[j].Type
	})
	sort.Ints(p.DeclaredPorts)
	sort.Strings(p.EnvironmentNames)
	sort.Strings(p.ComponentCatalogIDs)
	sort.Strings(p.BootstrapFieldNames)
	return p
}

func (p RepoProfile) Validate() error {
	if p.SchemaVersion != RepoProfileSchemaVersion || strings.TrimSpace(p.TenantID) == "" || strings.TrimSpace(p.ProjectID) == "" {
		return errors.New("invalid RepoProfile identity or schema")
	}
	for _, manifest := range p.Manifests {
		if strings.TrimSpace(manifest.Path) == "" || strings.TrimSpace(manifest.Type) == "" || strings.ContainsAny(manifest.Path, "\r\n") {
			return errors.New("invalid RepoProfile manifest")
		}
	}
	for _, port := range p.DeclaredPorts {
		if port < 1 || port > 65535 {
			return errors.New("invalid RepoProfile port")
		}
	}
	for _, name := range append(append([]string{}, p.EnvironmentNames...), append(p.ComponentCatalogIDs, p.BootstrapFieldNames...)...) {
		if strings.TrimSpace(name) == "" || len(name) > 128 || strings.ContainsAny(name, "\r\n") {
			return errors.New("invalid RepoProfile identifier")
		}
	}
	return nil
}
