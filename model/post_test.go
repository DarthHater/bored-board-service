package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPost(t *testing.T) {
	m := Post{}
	m.Body = "Test"
	assert.Equal(t, m.Body, "Test")
}
