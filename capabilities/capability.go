// Package capabilities defines the abstract capability interfaces that form the
// contract between the CloudOS kernel and every provider. Providers implement
// capabilities; the kernel discovers and orchestrates them.
//
// This package must have zero dependencies on kernel, provider, or application
// code. It depends only on the standard library and packages/types.
package capabilities

import (
	"context"
	"fmt"

	"github.com/cloudos/cloudos/packages/types"
)

// ID is a unique, well-known identifier for a capability (e.g. "compute", "storage").
type ID string

// Version carries the semantic version of a capability interface.
// Providers advertise which version of a capability they implement.
type Version struct {
	Major int `json:"major" yaml:"major"`
	Minor int `json:"minor" yaml:"minor"`
	Patch int `json:"patch" yaml:"patch"`
}

// String returns the dotted version string.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Capability is the base interface that every capability must implement.
// It provides identification, versioning, and lifecycle methods.
type Capability interface {
	// ID returns the well-known capability identifier.
	ID() ID

	// Version returns the interface version this capability implements.
	Version() Version

	// Validate checks whether the capability is in a valid state.
	Validate(ctx context.Context) error
}

// Compute defines the interface for compute / deployment capabilities.
type Compute interface {
	Capability
	Deploy(ctx context.Context, req DeployRequest) (*Deployment, error)
	GetDeployment(ctx context.Context, id types.ResourceID) (*Deployment, error)
	ListDeployments(ctx context.Context) ([]*Deployment, error)
	RemoveDeployment(ctx context.Context, id types.ResourceID) error
	Exec(ctx context.Context, id types.ResourceID, cmd []string) ([]byte, error)
	Logs(ctx context.Context, id types.ResourceID, tail int) ([]string, error)
}

// Storage defines the interface for object / file storage capabilities.
type Storage interface {
	Capability
	Put(ctx context.Context, bucket string, key string, data []byte, contentType string) error
	Get(ctx context.Context, bucket string, key string) ([]byte, error)
	Delete(ctx context.Context, bucket string, key string) error
	List(ctx context.Context, bucket string, prefix string) ([]*Object, error)
	CreateBucket(ctx context.Context, bucket string) error
	DeleteBucket(ctx context.Context, bucket string) error
}

// Database defines the interface for database / SQL capabilities.
type Database interface {
	Capability
	Exec(ctx context.Context, query string, args ...interface{}) (Result, error)
	Query(ctx context.Context, query string, args ...interface{}) (Rows, error)
	Migrate(ctx context.Context, migrations []Migration) error
	Ping(ctx context.Context) error
}

// AI defines the interface for AI / LLM capabilities.
type AI interface {
	Capability
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	Embed(ctx context.Context, input []string) ([][]float32, error)
	ListModels(ctx context.Context) ([]*ModelInfo, error)
	Stream(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)
}

// Network defines the interface for networking / DNS capabilities.
type Network interface {
	Capability
	AllocateIP(ctx context.Context) (string, error)
	ReleaseIP(ctx context.Context, ip string) error
	CreateNetwork(ctx context.Context, req NetworkRequest) (*NetworkResponse, error)
	DeleteNetwork(ctx context.Context, id string) error
	ResolveDNS(ctx context.Context, name string) ([]string, error)
}

// --- Request / Response types -------------------------------------------------

// DeployRequest carries the parameters for creating a deployment.
type DeployRequest struct {
	Name     string            `json:"name"`
	Image    string            `json:"image"`
	Port     int               `json:"port"`
	Replicas int               `json:"replicas"`
	Env      map[string]string `json:"env,omitempty"`
}

// Deployment represents a running deployment.
type Deployment struct {
	ID        types.ResourceID    `json:"id"`
	Name      string              `json:"name"`
	Image     string              `json:"image"`
	Status    types.ResourceState `json:"status"`
	Port      int                 `json:"port"`
	Replicas  int                 `json:"replicas"`
	CreatedAt int64               `json:"createdAt"`
}

// Object represents a stored object.
type Object struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ContentType  string `json:"contentType"`
	LastModified int64  `json:"lastModified"`
}

// Result holds the outcome of a database Exec call.
type Result struct {
	RowsAffected int64 `json:"rowsAffected"`
	LastInsertID int64 `json:"lastInsertID"`
}

// Rows holds the outcome of a database Query call.
type Rows struct {
	Columns []string          `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
}

// Migration describes a single database migration step.
type Migration struct {
	ID      string `json:"id"`
	Query   string `json:"query"`
	Description string `json:"description"`
}

// ChatRequest carries the parameters for an AI chat completion call.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message is a single message in a chat conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse carries the result of an AI chat completion call.
type ChatResponse struct {
	Message Message `json:"message"`
	Usage   Usage   `json:"usage"`
}

// ChatChunk is a streaming chunk from an AI chat completion.
type ChatChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
}

// Usage carries token usage information.
type Usage struct {
	PromptTokens     int `json:"promptTokens"`
	CompletionTokens int `json:"completionTokens"`
	TotalTokens      int `json:"totalTokens"`
}

// ModelInfo describes an available AI model.
type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NetworkRequest carries the parameters for creating a network.
type NetworkRequest struct {
	Name   string `json:"name"`
	CIDR   string `json:"cidr"`
	Region string `json:"region,omitempty"`
}

// NetworkResponse carries the result of a network creation.
type NetworkResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	CIDR    string `json:"cidr"`
	Status  string `json:"status"`
	Created int64  `json:"created"`
}
