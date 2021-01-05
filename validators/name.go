package validators

import (
	"strings"
)

const (
	MaxUsernameLength = 30
	MaxNameLength     = 40

	//TODO take more symbols
	InvalidSym = "@?.,&^%$#@!^&*()-+=:''`?.,"
)

//TODO transfer to any pkg
func IsValidUsername(username string) bool {
	if len(username) > MaxUsernameLength {
		return false
	}

	count := strings.Count(" ", username)
	//TODO test
	for _, v := range InvalidSym {
		count += strings.Count(string(v), username)
	}

	return count == 0
}

//TODO  transfer to any pkg
func IsValidName(name string) bool {
	if len(name) > MaxNameLength {
		return false
	}

	var count int
	//TODO test
	for _, v := range InvalidSym {
		count += strings.Count(string(v), name)
	}

	return count == 0
}
