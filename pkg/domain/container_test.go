package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetailsLoaded(t *testing.T) {
	t.Parallel()

	notLoaded := &Container{ID: "abc"}
	assert.False(t, notLoaded.DetailsLoaded(), "a container with nil Details is not loaded")

	loaded := &Container{ID: "abc", Details: &ContainerDetails{}}
	assert.True(t, loaded.DetailsLoaded(), "a container with non-nil Details is loaded")
}
