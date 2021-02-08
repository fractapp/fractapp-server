package validators

import (
	"regexp"
)

const (
	MaxUsernameLength = 30
	MinUsernameLength = 4
	MaxNameLength     = 32
	MinNameLength     = 4

	patternForUsername = "^[0-9a-z]*$"
	patternForName     = "^[0-9a-zA-z ]*$"
)

func IsValidUsername(username string) bool {
	if len(username) > MaxUsernameLength || len(username) < MinUsernameLength {
		return false
	}

	if v, _ := regexp.MatchString(patternForUsername, username); !v {
		return false
	}

	return true
}

func IsValidName(name string) bool {
	if len(name) > MaxNameLength || len(name) < MinNameLength {
		return false
	}

	if v, _ := regexp.MatchString(patternForUsername, patternForName); !v {
		return false
	}

	return true
}