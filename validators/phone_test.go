package validators

import (
	"testing"

	"gotest.tools/assert"
)

func TestIsValidatePhoneNumberPositive(t *testing.T) {
	assert.Assert(t, IsValidatePhoneNumber("+12025550180"))
}
func TestIsValidatePhoneNumberMatchRegExp(t *testing.T) {
	assert.Assert(t, !IsValidatePhoneNumber("+202-555-0180"))
}
func TestIsValidatePhoneNumberMaxLength(t *testing.T) {
	assert.Assert(t, !IsValidatePhoneNumber("1234512345123451"))
}
func TestIsValidatePhoneNumberWithoutPlus(t *testing.T) {
	assert.Assert(t, !IsValidatePhoneNumber("12025550180"))
}
