package domain

import (
	"strings"
	"time"
)

// Provider identifies the source control provider for a normalized event.
type Provider string

const (
	ProviderGitHub Provider = "github"
	ProviderGitLab Provider = "gitlab"
)

// EventAction is the lifecycle action represented by a normalized event.
type EventAction string

const (
	ActionOpen   EventAction = "open"
	ActionUpdate EventAction = "update"
	ActionClose  EventAction = "close"
	ActionIgnore EventAction = "ignore"
)

type PRCommand string

const (
	CommandRecreate PRCommand = "recreate"
	CommandDestroy  PRCommand = "destroy"
	CommandPin      PRCommand = "pin"
)

// PullRequestEvent is the versioned webhook-to-control-plane contract.
type PullRequestEvent struct {
	Provider       Provider    `json:"provider"`
	Action         EventAction `json:"action"`
	Repo           string      `json:"repo"`
	Branch         string      `json:"branch"`
	ChangeID       string      `json:"changeId"`
	CommitSHA      string      `json:"commitSha"`
	Author         string      `json:"author"`
	URL            string      `json:"url"`
	EventID        string      `json:"eventId"`
	InstallationID string      `json:"installationId"`
	Labels         []string    `json:"labels"`
	Draft          bool        `json:"draft"`
	ArtifactDigest string      `json:"artifactDigest,omitempty"`
	Sequence       int64       `json:"sequence,omitempty"`
}

type PullRequestCommand struct {
	Provider       Provider      `json:"provider"`
	Command        PRCommand     `json:"command"`
	Repo           string        `json:"repo"`
	ChangeID       string        `json:"changeId"`
	Author         string        `json:"author"`
	URL            string        `json:"url"`
	EventID        string        `json:"eventId"`
	InstallationID string        `json:"installationId"`
	PinDuration    time.Duration `json:"pinDuration,omitempty"`
	PinRaw         string        `json:"pinRaw,omitempty"`
}

func (e PullRequestEvent) EnvironmentID() string {
	switch e.Provider {
	case ProviderGitHub:
		return "pr-" + normalizeSCMIdentifier(e.ChangeID)
	case ProviderGitLab:
		return "mr-" + normalizeSCMIdentifier(e.ChangeID)
	default:
		if id := normalizeSCMIdentifier(e.ChangeID); id != "" {
			return id
		}
		return normalizeSCMIdentifier(e.Branch)
	}
}

func (c PullRequestCommand) EnvironmentID() string {
	return PullRequestEvent{Provider: c.Provider, ChangeID: c.ChangeID}.EnvironmentID()
}

func (c PullRequestCommand) PullRequestEvent(action EventAction) PullRequestEvent {
	return PullRequestEvent{Provider: c.Provider, Action: action, Repo: c.Repo, ChangeID: c.ChangeID, Author: c.Author, URL: c.URL, EventID: c.EventID, InstallationID: c.InstallationID}
}

func (e PullRequestEvent) DeduplicationKey() string {
	if value := strings.TrimSpace(e.EventID); value != "" {
		return "event:" + strings.ToLower(value)
	}
	parts := []string{strings.ToLower(strings.TrimSpace(string(e.Provider))), strings.ToLower(strings.TrimSpace(e.Repo)), strings.ToLower(strings.TrimSpace(string(e.Action)))}
	if changeID := strings.TrimSpace(e.ChangeID); changeID != "" {
		parts = append(parts, changeID)
	} else if branch := strings.TrimSpace(e.Branch); branch != "" {
		parts = append(parts, strings.ToLower(branch))
	} else {
		return ""
	}
	if parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return ""
	}
	return strings.Join(parts, "|")
}

func (e PullRequestEvent) CreateEnvironmentRequest(product, project string) CreateEnvironmentRequest {
	if product == "" {
		product = "generic"
	}
	if project == "" {
		project = scmProjectName(e.Repo)
	}
	return CreateEnvironmentRequest{ID: e.EnvironmentID(), Project: project, Product: product, Mode: ModeFull, Source: SCMSource{Provider: string(e.Provider), Repository: e.Repo, PullRequestID: e.ChangeID, Branch: e.Branch, Commit: e.CommitSHA, Author: e.Author, URL: e.URL}, DesiredRevision: EnvironmentRevision{Provider: e.Provider, Repository: e.Repo, ChangeID: e.ChangeID, Commit: e.CommitSHA, ArtifactDigest: e.ArtifactDigest, Sequence: e.Sequence}}
}

func normalizeSCMIdentifier(value string) string {
	return NormalizeEnvironmentID(value)
}

func scmProjectName(repo string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(repo), func(r rune) bool { return r == '/' || r == ':' || r == '\\' })
	if len(parts) == 0 {
		return "default"
	}
	name := strings.TrimSuffix(normalizeSCMIdentifier(parts[len(parts)-1]), ".git")
	if name == "" {
		return "default"
	}
	return name
}
