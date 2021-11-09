package types

import (
	"math/big"
)

type Currency int32

const (
	DOT Currency = iota
	KSM
)

var Currencies = []Currency{
	DOT,
	KSM,
}

func (c Currency) ConvertFromPlanck(amount *big.Int) *big.Float {
	decimals := c.Decimals()

	d := new(big.Int)
	d.Exp(big.NewInt(10), big.NewInt(decimals), nil)

	return new(big.Float).Quo(new(big.Float).SetInt(amount), new(big.Float).SetInt(d))
}

func (c Currency) Accuracy() int64 {
	switch c {
	case DOT:
		return 1000
	case KSM:
		return 1000
	}

	return 1000
}

func (c Currency) Network() (n Network) {
	n = Polkadot

	switch c {
	case DOT:
		n = Polkadot
	case KSM:
		n = Kusama
	}

	return n
}

func (c Currency) ConvertFromPlanckToView(amount *big.Int) *big.Float {
	decimals := c.Decimals()

	d := new(big.Int)
	d.Exp(big.NewInt(10), big.NewInt(decimals), nil)

	return new(big.Float).Quo(new(big.Float).SetInt(amount), new(big.Float).SetInt(d))
}

func (c Currency) Decimals() int64 {
	switch c {
	case DOT:
		return 10
	case KSM:
		return 12
	}

	return 10
}

func (c Currency) String() string {
	switch c {
	case DOT:
		return "DOT"
	case KSM:
		return "KSM"
	}

	return "DOT"
}
