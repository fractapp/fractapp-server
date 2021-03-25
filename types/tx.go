package types

import "math/big"

type Tx struct {
	Sender     string
	Receiver   string
	FullAmount *big.Int
	Fee        *big.Int
	Timestamp  int64
}
