package validators

import "regexp"

const (
	phonePattern = "^[0-9]*$"
)

func IsValidatePhoneNumber(phoneNumber string) bool {
	if len(phoneNumber) > 15 {
		return false
	}
	if phoneNumber[0] != '+' {
		return false
	}
	if v, _ := regexp.MatchString(phonePattern, phoneNumber[1:]); !v {
		return false
	}

	return true
}
