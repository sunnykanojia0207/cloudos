// Package types defines the shared domain types used across all CloudOS packages.
// Every subsystem, capability, and provider references these base types.
package types

import "time"

// ResourceID is a unique identifier for any CloudOS resource.
// Encoded as a UUID string throughout the system.
type ResourceID string

// ResourceMeta contains common metadata embedded in every CloudOS resource.
type ResourceMeta struct {
	ID        ResourceID `json:"id" yaml:"id"`
	Name      string     `json:"name" yaml:"name"`
	CreatedAt time.Time  `json:"createdAt" yaml:"created_at"`
	UpdatedAt time.Time  `json:"updatedAt" yaml:"updated_at"`
}

// ResourceState represents the lifecycle state of a CloudOS resource.
type ResourceState string

const (
	StatePending  ResourceState = "pending"
	StateRunning  ResourceState = "running"
	StateStopped  ResourceState = "stopped"
	StateFailed   ResourceState = "failed"
	StateDegraded ResourceState = "degraded"
	StateUnknown  ResourceState = "unknown"
)

// HealthStatus represents the health check result of a subsystem or provider.
type HealthStatus struct {
	State     ResourceState `json:"state" yaml:"state"`
	Message   string        `json:"message,omitempty" yaml:"message,omitempty"`
	Timestamp time.Time     `json:"timestamp" yaml:"timestamp"`
}

// IsHealthy returns true when the health status indicates a working system.
func (h HealthStatus) IsHealthy() bool {
	return h.State == StateRunning
}

// Pagination carries pagination metadata for list-style API responses.
type Pagination struct {
	Page    int `json:"page" yaml:"page"`
	PerPage int `json:"perPage" yaml:"per_page"`
	Total   int `json:"total" yaml:"total"`
}

// ErrorDetail provides a structured error for API responses and cross-service error propagation.
type ErrorDetail struct {
	Code    string `json:"code" yaml:"code"`
	Message string `json:"message" yaml:"message"`
	Detail  string `json:"detail,omitempty" yaml:"detail,omitempty"`
}

// Error implements the error interface for ErrorDetail.
func (e ErrorDetail) Error() string {
	return e.Message
}
