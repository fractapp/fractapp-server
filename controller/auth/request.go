package auth

import (
	"fractapp-server/notification"
	"fractapp-server/types"
)

type TokenRs struct {
	Token string `json:"token"` // JWT token
}
type SendCodeRq struct {
	Type  notification.NotificatorType `enums:"0,1"` // Message type (0 - sms / 1 - email)
	Value string                       // Email address or Phone number (without +)
}
type ConfirmAuthRq struct {
	Value     string                       // Email address or Phone number (without +)
	Type      notification.NotificatorType `enums:"0,1,2"` // Message type with code (0 - sms / 1 - email)
	Addresses map[types.Network]Address    // Addresses by network (0 - polkadot/ 1 - kusama) from account
	Code      string                       // The code that was sent
}
type Address struct {
	Address string // Blockchain address from account
	PubKey  string // PubKey from account
	Sign    string // Sign for message (more information here: https://github.com/fractapp/fractapp-server/blob/main/AUTH.md)
}
