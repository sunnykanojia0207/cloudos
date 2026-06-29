// Package project implements the CloudOS Project resource — the primary
// workspace for users. Every user-facing resource (applications, deployments,
// databases, secrets, domains, AI agents) belongs to a Project.
//
// Projects are the heart of the CloudOS user experience. Instead of showing
// users a raw list of cloud services (like AWS), CloudOS shows:
//
//	My Projects
//	  ├── CRM
//	  ├── Portfolio Website
//	  ├── AI Assistant
//	  └── Inventory System
//
// Opening a project reveals its applications, deployments, databases, and more.
//
// This package provides the Project resource type (spec + status), validation,
// defaults, and the ProjectController that reconciles project state.
package project

import (
	"fmt"
	"regexp"
	"time"

	"github.com/cloudos/cloudos/kernel/resource"
)

// ── Constants ──────────────────────────────────────────────────────────────

const (
	// Kind is the resource kind string for Project.
	Kind = "Project"

	// Environment types.
	EnvDevelopment = "development"
	EnvStaging     = "staging"
	EnvProduction  = "production"

	// Phase values.
	PhaseActive     = "Active"
	PhaseArchived   = "Archived"
	PhaseDeleting   = "Deleting"
	PhaseCreating   = "Creating"

	// Health values.
	HealthHealthy  = "Healthy"
	HealthDegraded = "Degraded"
	HealthError    = "Error"
)

// ValidEnvironments is the set of allowed environment values.
var ValidEnvironments = map[string]bool{
	EnvDevelopment: true,
	EnvStaging:     true,
	EnvProduction:  true,
}

// projectIDPattern enforces DNS-label-compatible project IDs.
var projectIDPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// DefaultProjectSettings are the settings applied when a project is created.
var DefaultProjectSettings = map[string]string{
	"autoDeploy":       "true",
	"autoBackup":       "false",
	"monitoring":       "basic",
	"notifications":    "enabled",
	"versioning":       "enabled",
}

// ── Condition ──────────────────────────────────────────────────────────────

// Condition represents a single status condition for a resource.
// Inspired by Kubernetes condition types.
type Condition struct {
	// Type is the condition type (e.g. "Ready", "Initialized").
	Type string `json:"type"`

	// Status is one of "True", "False", "Unknown".
	Status string `json:"status"`

	// Reason is a machine-readable reason code.
	Reason string `json:"reason,omitempty"`

	// Message is a human-readable explanation.
	Message string `json:"message,omitempty"`

	// LastTransitionTime is when the condition last changed.
	LastTransitionTime time.Time `json:"lastTransitionTime,omitempty"`
}

// ── AI Settings ────────────────────────────────────────────────────────────

// AISettings configures AI-related features for a project.
type AISettings struct {
	// Enabled enables AI features for this project.
	Enabled bool `json:"enabled"`

	// DefaultModel is the default AI model to use.
	DefaultModel string `json:"defaultModel,omitempty"`

	// MaxTokensPerRequest limits token usage per AI request.
	MaxTokensPerRequest int `json:"maxTokensPerRequest,omitempty"`

	// AllowedProviders lists which AI providers are allowed.
	AllowedProviders []string `json:"allowedProviders,omitempty"`
}

// ── Quota ──────────────────────────────────────────────────────────────────

// QuotaSpec defines resource limits for a project.
type QuotaSpec struct {
	// MaxApplications is the maximum number of applications.
	MaxApplications int `json:"maxApplications,omitempty"`

	// MaxDeployments is the maximum number of deployments.
	MaxDeployments int `json:"maxDeployments,omitempty"`

	// MaxDatabases is the maximum number of databases.
	MaxDatabases int `json:"maxDatabases,omitempty"`

	// MaxStorageGB is the maximum storage in gigabytes.
	MaxStorageGB int `json:"maxStorageGB,omitempty"`

	// MaxSecrets is the maximum number of secrets.
	MaxSecrets int `json:"maxSecrets,omitempty"`

	// MaxDomains is the maximum number of custom domains.
	MaxDomains int `json:"maxDomains,omitempty"`
}

// ── ProjectSpec ────────────────────────────────────────────────────────────

// ProjectSpec is the desired state of a CloudOS Project.
type ProjectSpec struct {
	// DisplayName is a human-readable name for the project.
	DisplayName string `json:"displayName"`

	// Description explains the purpose of this project.
	Description string `json:"description,omitempty"`

	// Environment is the deployment environment.
	// One of: "development", "staging", "production".
	Environment string `json:"environment"`

	// DefaultRegion is the default cloud region for resources.
	DefaultRegion string `json:"defaultRegion,omitempty"`

	// DefaultProviders lists default provider IDs.
	DefaultProviders []string `json:"defaultProviders,omitempty"`

	// Tags are user-defined labels for categorization.
	Tags []string `json:"tags,omitempty"`

	// AISettings configures AI features.
	AISettings *AISettings `json:"aiSettings,omitempty"`

	// Quota limits resource usage for this project.
	Quota *QuotaSpec `json:"quota,omitempty"`

	// Settings are arbitrary key-value configuration pairs.
	Settings map[string]string `json:"settings,omitempty"`
}

// ── ProjectStatus ──────────────────────────────────────────────────────────

// ProjectStatus is the current observed state of a CloudOS Project.
type ProjectStatus struct {
	// Phase is the project lifecycle phase.
	// One of: "Creating", "Active", "Archived", "Deleting".
	Phase string `json:"phase"`

	// Health is the overall operational health.
	// One of: "Healthy", "Degraded", "Error".
	Health string `json:"health"`

	// Conditions provide detailed status signals.
	Conditions []Condition `json:"conditions,omitempty"`

	// LastActivity is when the project was last modified.
	LastActivity time.Time `json:"lastActivity,omitempty"`

	// ResourceCount is the number of resources in this project.
	ResourceCount int `json:"resourceCount"`

	// DeploymentCount is the number of active deployments.
	DeploymentCount int `json:"deploymentCount"`
}

// ── Project Resource ───────────────────────────────────────────────────────

// Project is the concrete CloudOS Project resource. It implements the
// resource.Resource interface and can be used with the Resource Engine.
type Project struct {
	Metadata_ *resource.Metadata `json:"metadata"`
	Spec_     ProjectSpec        `json:"spec"`
	Status_   ProjectStatus      `json:"status"`
}

// NewProject creates a new Project resource with sensible defaults.
// The project is placed in the default namespace and initialized with
// Creating phase and Pending health.
func NewProject(id, displayName, environment, description string) *Project {
	now := time.Now()

	// Default to development environment if not specified.
	if !ValidEnvironments[environment] {
		environment = EnvDevelopment
	}

	return &Project{
		Metadata_: &resource.Metadata{
			ID:              id,
			Name:            displayName,
			Namespace:       resource.NamespaceDefault,
			Kind:            Kind,
			APIVersion:      resource.APIVersion,
			Labels:          make(map[string]string),
			Annotations:     make(map[string]string),
			CreatedAt:       now,
			UpdatedAt:       now,
			ResourceVersion: 1,
		},
		Spec_: ProjectSpec{
			DisplayName: displayName,
			Description: description,
			Environment: environment,
			Settings:    copyMap(DefaultProjectSettings),
		},
		Status_: ProjectStatus{
			Phase:          PhaseCreating,
			Health:         HealthHealthy,
			Conditions:     []Condition{},
			LastActivity:   now,
			ResourceCount:  0,
			DeploymentCount: 0,
		},
	}
}

// ── Resource Interface ─────────────────────────────────────────────────────

func (p *Project) GetKind() string       { return Kind }
func (p *Project) GetMetadata() *resource.Metadata { return p.Metadata_ }
func (p *Project) GetSpec() interface{}  { return p.Spec_ }
func (p *Project) GetStatus() interface{} { return p.Status_ }

func (p *Project) SetStatus(s interface{}) {
	if st, ok := s.(ProjectStatus); ok {
		p.Status_ = st
	}
}

// Validate checks the Project for semantic correctness.
func (p *Project) Validate() error {
	if p.Metadata_.ID == "" {
		return fmt.Errorf("project id is required")
	}
	if !projectIDPattern.MatchString(p.Metadata_.ID) {
		return fmt.Errorf("project id %q must match %s (lowercase letters, digits, and hyphens; must start and end with alphanumeric)",
			p.Metadata_.ID, projectIDPattern.String())
	}
	if p.Spec_.DisplayName == "" {
		return fmt.Errorf("project display name is required")
	}
	if p.Spec_.Environment != "" && !ValidEnvironments[p.Spec_.Environment] {
		return fmt.Errorf("project environment %q is invalid; must be one of: development, staging, production",
			p.Spec_.Environment)
	}
	if p.Metadata_.Kind != Kind {
		return fmt.Errorf("kind must be %q, got %q", Kind, p.Metadata_.Kind)
	}
	return nil
}

// ── Defaults ───────────────────────────────────────────────────────────────

// EnsureDefaults populates any missing default values in the Project's spec
// and metadata. Unlike Validate(), this does not return errors — it fixes
// common omissions.
func (p *Project) EnsureDefaults() {
	if p.Metadata_.Labels == nil {
		p.Metadata_.Labels = make(map[string]string)
	}
	if p.Metadata_.Annotations == nil {
		p.Metadata_.Annotations = make(map[string]string)
	}
	// Set the environment label.
	p.Metadata_.Labels["environment"] = p.Spec_.Environment

	if p.Spec_.Settings == nil {
		p.Spec_.Settings = copyMap(DefaultProjectSettings)
	}
	if p.Status_.Phase == "" {
		p.Status_.Phase = PhaseCreating
	}
	if p.Status_.Health == "" {
		p.Status_.Health = HealthHealthy
	}
	if p.Status_.Conditions == nil {
		p.Status_.Conditions = []Condition{}
	}
}

// ── Helpers ────────────────────────────────────────────────────────────────

// AddCondition adds or updates a status condition.
func (p *Project) AddCondition(condType, status, reason, message string) {
	now := time.Now()
	for i, c := range p.Status_.Conditions {
		if c.Type == condType {
			p.Status_.Conditions[i].Status = status
			p.Status_.Conditions[i].Reason = reason
			p.Status_.Conditions[i].Message = message
			p.Status_.Conditions[i].LastTransitionTime = now
			return
		}
	}
	p.Status_.Conditions = append(p.Status_.Conditions, Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: now,
	})
}

// GetCondition returns a condition by type.
func (p *Project) GetCondition(condType string) *Condition {
	for _, c := range p.Status_.Conditions {
		if c.Type == condType {
			return &c
		}
	}
	return nil
}

// Touch updates the LastActivity timestamp and increments the resource version.
func (p *Project) Touch() {
	p.Status_.LastActivity = time.Now()
}

// copyMap creates a shallow copy of a string map.
func copyMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
