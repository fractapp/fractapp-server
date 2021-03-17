package notification

import "fractapp-server/types"

type UpdateTokenRq struct {
	PubKey    string        // Public key from address
	Address   string        // Blockchain address
	Network   types.Network // network id (0 - polkadot/ 1 - kusama) from address
	Sign      string        // signature for message (more information here: https://github.com/fractapp/fractapp-server/blob/main/AUTH.md)
	Token     string        // firebase token
	Timestamp int64         // timestamp from message
}
