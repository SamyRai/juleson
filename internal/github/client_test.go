package github

import (
	"testing"

	"github.com/SamyRai/go-jules"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		client := NewClient("", nil)
		assert.Nil(t, client)
	})

	t.Run("valid token", func(t *testing.T) {
		// Mock jules client can be nil for this test since NewClient doesn't dereference it
		// during initialization.
		var julesClient *jules.Client

		client := NewClient("dummy_token", julesClient)

		assert.NotNil(t, client)
		assert.NotNil(t, client.Client)
		assert.Equal(t, "dummy_token", client.token)

		// Check that all services are initialized
		assert.NotNil(t, client.Repositories)
		assert.NotNil(t, client.Actions)
		assert.NotNil(t, client.PullRequests)
		assert.NotNil(t, client.Sessions)
		assert.NotNil(t, client.Issues)
		assert.NotNil(t, client.Milestones)
		assert.NotNil(t, client.Projects)
	})
}
