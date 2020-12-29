package controller

import (
	"errors"
	"time"
)

const (
	SignTimeout = 10 * time.Minute
)

var (
	InvalidAddressErr  = errors.New("address not equals pubkey")
	InvalidSignTimeErr = errors.New("invalid sign time")
	InvalidAuthErr     = errors.New("invalid auth")
	InvalidRqErr       = errors.New("invalid rq")
)
