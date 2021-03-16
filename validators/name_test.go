package validators

import (
	"testing"

	"gotest.tools/assert"
)

func TestIsValidUsernamePositive(t *testing.T) {
	assert.Assert(t, IsValidUsername("testname"))
}
func TestIsValidUsernameMaxUsernameLength(t *testing.T) {
	assert.Assert(t, !IsValidUsername("1111111111111111111111111111111"))
}
func TestIsValidUsernameMinUsernameLength(t *testing.T) {
	assert.Assert(t, !IsValidUsername("tes"))
}
func TestIsValidUsernameMatchRegExp(t *testing.T) {
	assert.Assert(t, !IsValidUsername("test123@"))
}

func TestIsValidNamePositive(t *testing.T) {
	assert.Assert(t, IsValidName("Test Boy"))
}
func TestIsValidNameMaxUsernameLength(t *testing.T) {
	assert.Assert(t, !IsValidName("111111111 11111111111111111111111"))
}
func TestIsValidNameMinUsernameLength(t *testing.T) {
	assert.Assert(t, !IsValidName("tes"))
}
func TestIsValidNameMatchRegExp(t *testing.T) {
	assert.Assert(t, !IsValidName("Test test123@"))
}
