package types

import (
	"math/big"
	"testing"
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
