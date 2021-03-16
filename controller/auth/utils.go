package auth

import (
	"fmt"
	"math/rand"
	"time"
)

func generateCode() string {
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	codeInt := generator.Intn(999999)
	code := fmt.Sprintf("%d", codeInt)
	if len(code) < 6 {
		code = fmt.Sprintf("%06s", code)
	}

	return code
}
