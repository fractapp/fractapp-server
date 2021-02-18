package auth

import (
	"fractapp-server/notification"
	"fractapp-server/types"
)

type TokenRs struct {
	Token string `json:"token"`
}
type SendCodeRq struct {
	Type      notification.NotificatorType
	CheckType notification.CheckType
	Value     string
}
type ConfirmRegRq struct {
	Value     string
	Type      notification.NotificatorType
	Addresses map[types.Network]Address
	Code      string
}
type Address struct {
	Address string
	PubKey  string
	Sign    string
}
