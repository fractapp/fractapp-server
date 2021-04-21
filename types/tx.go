package types

import "math/big"

type AdaptorTx struct {
	EventID    string
	Sender     string
	Receiver   string
	FullAmount *big.Int
	Fee        *big.Int
	Timestamp  int64
}
