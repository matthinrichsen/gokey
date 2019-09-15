package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveQuotes(t *testing.T) {
	assert.Equal(t, ``, RemoveQuotes(`""`))
	assert.Equal(t, `a`, RemoveQuotes(`"a"`))
	assert.Equal(t, `a"a`, RemoveQuotes(`"a"a"`))
}
