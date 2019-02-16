package rlimiter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInStrings(t *testing.T) {
	assert.True(t, inStrings("POST", []string{"GET", "POST"}))
	assert.False(t, inStrings("GET", []string{"PUT", "POST"}))
}
