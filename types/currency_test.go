package types

import (
	"math/big"
	"testing"

	"gotest.tools/assert"
)

func TestDOTToString(t *testing.T) {
	c := DOT
	if c.String() != "DOT" {
		t.Fatal()
	}
}

func TestKSMToString(t *testing.T) {
	c := KSM
	if c.String() != "KSM" {
		t.Fatal()
	}
}

func TestDefaultToString(t *testing.T) {
	c := Currency(999999)
	if c.String() != "DOT" {
		t.Fatal()
	}
}

func TestDOTDecimals(t *testing.T) {
	c := DOT
	if c.Decimals() != 10 {
		t.Fatal()
	}
}

func TestKSMDecimals(t *testing.T) {
	c := KSM
	if c.Decimals() != 12 {
		t.Fatal()
	}
}

func TestDefaultDecimals(t *testing.T) {
	c := Currency(999999)
	if c.Decimals() != 10 {
		t.Fatal()
	}
}

func TestConvertFromPlanck(t *testing.T) {
	c := DOT

	amount := big.NewInt(0)
	amount.SetString("12567899900000", 10)

	f := new(big.Float)
	f, _ = f.SetString("1256.78999")
	if c.ConvertFromPlanck(amount).Cmp(f) != 0 {
		t.Fatal()
	}
}

func TestAccuracy(t *testing.T) {
	assert.Equal(t, DOT.Accuracy(), int64(1000))
	assert.Equal(t, KSM.Accuracy(), int64(1000))
	assert.Equal(t, Currency(10000).Accuracy(), int64(1000))
}

func TestNetwork(t *testing.T) {
	assert.Equal(t, DOT.Network(), Polkadot)
	assert.Equal(t, KSM.Network(), Kusama)
	assert.Equal(t, Currency(10000).Network(), Polkadot)
}

func TestConvertFromPlanckToView(t *testing.T) {
	a := &big.Int{}
	a.SetString("10000010000", 10)

	assert.Equal(t, DOT.ConvertFromPlanckToView(a).String(), "1.000001")
	assert.Equal(t, KSM.ConvertFromPlanckToView(a).String(), "0.01000001")
	assert.Equal(t, Currency(10000).ConvertFromPlanckToView(a).String(), "1.000001")
}
