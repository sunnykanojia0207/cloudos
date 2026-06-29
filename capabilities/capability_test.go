package capabilities

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionString(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	assert.Equal(t, "1.2.3", v.String())
}

func TestVersionZero(t *testing.T) {
	v := Version{}
	assert.Equal(t, "0.0.0", v.String())
}

func TestDeploymentStruct(t *testing.T) {
	d := Deployment{Name: "test", Image: "nginx"}
	assert.Equal(t, "test", d.Name)
	assert.Equal(t, "nginx", d.Image)
}

func TestChatRequestStruct(t *testing.T) {
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	}
	assert.Equal(t, "gpt-4", req.Model)
	assert.Len(t, req.Messages, 1)
}
