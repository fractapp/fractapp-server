package auth

import (
	"testing"

	"gotest.tools/assert"
)

func TestGenerateCode(t *testing.T) {
	code := generateCode()
	twoCode := generateCode()

	assert.Assert(t, len(code) == 6)
	assert.Assert(t, len(twoCode) == 6)
	assert.Assert(t, code != twoCode)
}
