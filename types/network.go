package types

import (
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/blake2b"
)

type Network int

const (
	Polkadot Network = iota
	Kusama

	polkadotPrefix = byte(0)
	kusamaPrefix   = byte(2)
)

var (
	SS58prefix = []byte("SS58PRE")
)

func (n Network) Currency() Currency {
	switch n {
	case Polkadot:
		return DOT
	case Kusama:
		return KSM
	}

	return DOT
}

func (n Network) String() string {
	switch n {
	case Polkadot:
		return "polkadot"
	case Kusama:
		return "kusama"
	}

	return "polkadot"
}

func NetworkByCurrency(currency Currency) (n Network) {
	n = Polkadot

	switch currency {
	case DOT:
		n = Polkadot
	case KSM:
		n = Kusama
	}

	return n
}

func ParseNetwork(name string) (n Network) {
	n = Polkadot

	switch strings.ToLower(name) {
	case "polkadot":
		n = Polkadot
	case "kusama":
		n = Kusama
	}

	return n
}

func (n Network) StringToAddress(value string) []byte {
	var address []byte
	switch n {
	case Kusama:
		fallthrough
	case Polkadot:
		address = base58.Decode(value)
	}

	return address
}

func (n Network) Address(pubKey []byte) string {
	var address []byte
	switch n {
	case Polkadot:
		address = append([]byte{polkadotPrefix}, pubKey[:]...)
	case Kusama:
		address = append([]byte{kusamaPrefix}, pubKey[:]...)
	}

	hash := blake2b.Sum512(append(SS58prefix, address...))
	return base58.Encode(append(address, hash[0:2]...))
}
