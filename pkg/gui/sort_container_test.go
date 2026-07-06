package gui

import (
	"sort"
	"testing"

	"github.com/jesseduffield/lazydocker/pkg/domain"
	"github.com/stretchr/testify/assert"
)

func sampleContainers() []*domain.Container {
	return []*domain.Container{
		{
			ID:     "1",
			Name:   "1",
			Status: domain.StatusExited,
		},
		{
			ID:     "2",
			Name:   "2",
			Status: domain.StatusRunning,
		},
		{
			ID:     "3",
			Name:   "3",
			Status: domain.StatusRunning,
		},
		{
			ID:     "4",
			Name:   "4",
			Status: domain.StatusCreated,
		},
	}
}

func expectedPerStatusContainers() []*domain.Container {
	return []*domain.Container{
		{
			ID:     "2",
			Name:   "2",
			Status: domain.StatusRunning,
		},
		{
			ID:     "3",
			Name:   "3",
			Status: domain.StatusRunning,
		},
		{
			ID:     "1",
			Name:   "1",
			Status: domain.StatusExited,
		},
		{
			ID:     "4",
			Name:   "4",
			Status: domain.StatusCreated,
		},
	}
}

func expectedLegacySortedContainers() []*domain.Container {
	return []*domain.Container{
		{
			ID:     "1",
			Name:   "1",
			Status: domain.StatusExited,
		},
		{
			ID:     "2",
			Name:   "2",
			Status: domain.StatusRunning,
		},
		{
			ID:     "3",
			Name:   "3",
			Status: domain.StatusRunning,
		},
		{
			ID:     "4",
			Name:   "4",
			Status: domain.StatusCreated,
		},
	}
}

func assertEqualContainers(t *testing.T, left *domain.Container, right *domain.Container) {
	t.Helper()
	assert.Equal(t, left.Status, right.Status)
	assert.Equal(t, left.ID, right.ID)
	assert.Equal(t, left.Name, right.Name)
}

func TestSortContainers(t *testing.T) {
	actual := sampleContainers()

	expected := expectedPerStatusContainers()

	sort.Slice(actual, func(i, j int) bool {
		return sortContainers(actual[i], actual[j], false)
	})

	assert.Equal(t, len(actual), len(expected))

	for i := 0; i < len(actual); i++ {
		assertEqualContainers(t, expected[i], actual[i])
	}
}

func TestLegacySortedContainers(t *testing.T) {
	actual := sampleContainers()

	expected := expectedLegacySortedContainers()

	sort.Slice(actual, func(i, j int) bool {
		return sortContainers(actual[i], actual[j], true)
	})

	assert.Equal(t, len(actual), len(expected))

	for i := 0; i < len(actual); i++ {
		assertEqualContainers(t, expected[i], actual[i])
	}
}
