package types

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"gotest.tools/assert"
)

func TestNetworkToDOT(t *testing.T) {
	assert.Equal(t, Polkadot.Currency(), DOT)
}

func TestNetworkToKSM(t *testing.T) {
	assert.Equal(t, Kusama.Currency(), KSM)
}

func TestNetworkToDefault(t *testing.T) {
	assert.Equal(t, Network(999999).Currency(), DOT)
}

func TestPolkadotToString(t *testing.T) {
	assert.Equal(t, Polkadot.String(), "polkadot")
}

func TestKusamaToString(t *testing.T) {
	assert.Equal(t, Kusama.String(), "kusama")
}

func TestDefaultNetworkToString(t *testing.T) {
	assert.Equal(t, Network(999999).String(), "polkadot")
}

func TestParsePolkadot(t *testing.T) {
	assert.Equal(t, Polkadot, ParseNetwork("polkadot"))
}

func TestParseKusama(t *testing.T) {
	assert.Equal(t, Kusama, ParseNetwork("kusama"))
}

func TestParseDefault(t *testing.T) {
	assert.Equal(t, Polkadot, ParseNetwork("123123"))
}

func TestPolkadotAddress(t *testing.T) {
	pubKey, err := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Polkadot.Address(pubKey), "111111111111111111111111111111111HC1")
}

func TestKusamaAddress(t *testing.T) {
	pubKey, err := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, Kusama.Address(pubKey), "CaKWz5omakTK7ovp4m3koXrHyHb7NG3Nt7GENHbviByZpKp")
}
