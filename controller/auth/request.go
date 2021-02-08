package auth

import "fractapp-server/types"

type TokenRs struct {
	Token string `json:"token"`
}
type SendCodeRq struct {
	Type      types.CodeType
	CheckType types.CheckType
	Value     string
}
type ConfirmRegRq struct {
	Value     string
	Type      types.CodeType
	Addresses map[types.Network]Address
	Code      string
}
type Address struct {
	Address string
	PubKey  string
	Sign    string
}
