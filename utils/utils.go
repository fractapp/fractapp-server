package utils

import (
	"errors"

	"regexp"

	"github.com/ChainSafe/go-schnorrkel"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var (
	phonePattern   = "^[0-9]*$"
	InvalidSignErr = errors.New("invalid sign")
)

func Verify(pubKey [32]byte, msg string, hexSign string) error {
	srPublicKey := &(schnorrkel.PublicKey{})
	err := srPublicKey.Decode(pubKey)
	if err != nil {
		return err
	}

	signBytes, err := hexutil.Decode(hexSign)
	if err != nil {
		return InvalidSignErr
	}

	signingContext := schnorrkel.NewSigningContext([]byte("substrate"), []byte(msg))

	sign := [64]byte{}
	copy(sign[:], signBytes)

	signature := &(schnorrkel.Signature{})
	err = signature.Decode(sign)
	if err != nil {
		return InvalidSignErr
	}

	if !srPublicKey.Verify(signature, signingContext) {
		return InvalidSignErr
	}

	return nil
}
func Sign(privKey [32]byte, msg []byte) ([]byte, error) {
	miniSecretKey, err := schnorrkel.NewMiniSecretKeyFromRaw(privKey)
	if err != nil {
		return nil, err
	}
	secretKey := miniSecretKey.ExpandEd25519()
	signingContext := schnorrkel.NewSigningContext([]byte("substrate"), msg)

	sig, err := secretKey.Sign(signingContext)
	if err != nil {
		return nil, err
	}

	sigBytes := sig.Encode()
	return sigBytes[:], nil
}
func ParsePubKey(hex string) ([32]byte, error) {
	pubKey := [32]byte{}

	pubKeyBytes, err := hexutil.Decode(hex)
	if err != nil {
		return pubKey, err
	}
	copy(pubKey[:], pubKeyBytes)

	return pubKey, nil
}

func ValidatePhoneNumber(phoneNumber string) bool {
	if len(phoneNumber) > 15 {
		return false
	}
	if phoneNumber[0] != '+' {
		return false
	}
	if v, _ := regexp.MatchString(phonePattern, phoneNumber[1:]); !v {
		return false
	}

	return true
}
