package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultsWithAssert(t *testing.T) {
	assert.NotEmpty(t, Version)
	assert.NotEmpty(t, Commit)
	assert.NotEmpty(t, BuildDate)
}
