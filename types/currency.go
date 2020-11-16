package types

import "math/big"

type Currency int

const (
	accuracy          = 1000
	DOT      Currency = iota
	KSM
)

func (c Currency) ConvertFromPlanck(amount *big.Int) float64 {
	decimals := c.Decimals()

	amount.Mul(amount, big.NewInt(accuracy)).
		Div(amount, big.NewInt(decimals).Exp(big.NewInt(10), big.NewInt(decimals), nil))
	return float64(amount.Int64()) / accuracy
}

func (c Currency) Decimals() int64 {
	switch c {
	case DOT:
		return 10
	case KSM:
		return 12
	}

	return 0
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
