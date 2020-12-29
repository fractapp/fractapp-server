package profile

import "fractapp-server/types"

type ConfirmRegRq struct {
	PhoneNumber string
	Addresses   []Address
	Code        int
}
type Address struct {
	Address string
	PubKey  string
	Network types.Network
	Sign    string
}
type UpdateProfileRq struct {
	Name     string
	Username string
}
