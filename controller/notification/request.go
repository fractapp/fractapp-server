package notification

import "fractapp-server/types"

type UpdateTokenRq struct {
	PubKey    string
	Address   string
	Network   types.Network
	Sign      string
	Token     string
	Timestamp int64
}
