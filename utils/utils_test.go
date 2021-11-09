package utils

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"bou.ke/monkey"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"gotest.tools/assert"
)

func TestParsePubKeyPositive(t *testing.T) {
	validPub, err := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000")
	assert.Assert(t, err == nil)

	pub, err := ParsePubKey("0x0000000000000000000000000000000000000000000000000000000000000000")
	assert.Assert(t, err == nil)
	assert.Assert(t, bytes.Compare(pub[:], validPub) == 0)
}
func TestParsePubKeyNegative(t *testing.T) {
	_, err := ParsePubKey("123123")
	assert.Assert(t, err == hexutil.ErrMissingPrefix)
}

func TestVerifyPositive(t *testing.T) {
	privKeyBytes, err := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	var privKey [32]byte
	copy(privKey[:], privKeyBytes)

	pubKeyBytes, err := hexutil.Decode("0xdef12e42f3e487e9b14095aa8d5cc16a33491f1b50dadcf8811d1480f3fa8627")
	if err != nil {
		t.Fatal(err)
	}
	var pubKey [32]byte
	copy(pubKey[:], pubKeyBytes)

	msg := "test msg positive"
	err = Verify(pubKey, msg, "0xc4f20c3c6fab67a72ec02664f9a33b4f087b36fd24ce807a5f8652ba5bcf9e6c2bd443206057ba5c6efc779216253483b30427c32f5a68cc87b7ce495cacf385")
	assert.Assert(t, err == nil)
}
func TestVerifyNegative(t *testing.T) {
	privKeyBytes, err := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	var privKey [32]byte
	copy(privKey[:], privKeyBytes)

	pubKeyBytes, err := hexutil.Decode("0xdef12e42f3e487e9b14095aa8d5cc16a33491f1b50dadcf8811d1480f3fa8627")
	if err != nil {
		t.Fatal(err)
	}
	var pubKey [32]byte
	copy(pubKey[:], pubKeyBytes)

	msg := "test msg positive"

	err = Verify(pubKey, msg, "0x0000000000000000000000000000000000000000000000000000000000000000")
	assert.Assert(t, err == InvalidSignErr)
}
func TestSignPositive(t *testing.T) {
	privKeyBytes, err := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	var privKey [32]byte
	copy(privKey[:], privKeyBytes)

	pubKeyBytes, err := hexutil.Decode("0xdef12e42f3e487e9b14095aa8d5cc16a33491f1b50dadcf8811d1480f3fa8627")
	if err != nil {
		t.Fatal(err)
	}
	var pubKey [32]byte
	copy(pubKey[:], pubKeyBytes)

	msg := "test msg"
	sign, err := Sign(privKey, []byte(msg))
	if err != nil {
		t.Fatal(err)
	}

	err = Verify(pubKey, msg, hexutil.Encode(sign))
	assert.Assert(t, err == nil)
}
func TestVerifyInvalidSign(t *testing.T) {
	privKeyBytes, err := hexutil.Decode("0x0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	var privKey [32]byte
	copy(privKey[:], privKeyBytes)

	pubKeyBytes, err := hexutil.Decode("0xdef12e42f3e487e9b14095aa8d5cc16a33491f1b50dadcf8811d1480f3fa8627")
	if err != nil {
		t.Fatal(err)
	}
	var pubKey [32]byte
	copy(pubKey[:], pubKeyBytes)

	msg := "test msg"

	err = Verify(pubKey, msg, "123123")
	assert.Assert(t, err == InvalidSignErr)
}

func TestRandomHex(t *testing.T) {
	v, err := RandomHex(10)
	assert.Assert(t, err == nil)
	assert.Equal(t, len(v), 20)
}

func TestWriteAvatar(t *testing.T) {
	filename := "file-avatar"
	v := []byte{byte(1), byte(2), byte(5), byte(7), byte(8)}

	nameFn := ""
	flagFn := 0
	var permFn os.FileMode

	f := &os.File{}
	openFilePatch := monkey.Patch(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
		nameFn = name
		flagFn = flag
		permFn = perm
		return f, nil
	})
	defer openFilePatch.Unpatch()

	sizeFn := int64(100)
	truncateMock := monkey.PatchInstanceMethod(reflect.TypeOf(f), "Truncate", func(file *os.File, size int64) error {
		sizeFn = size
		return nil
	})
	defer truncateMock.Unpatch()

	var bFn []byte
	writeMock := monkey.PatchInstanceMethod(reflect.TypeOf(f), "Write", func(file *os.File, b []byte) (n int, err error) {
		bFn = b
		return 0, nil
	})
	defer writeMock.Unpatch()

	err := WriteAvatar(filename, v)
	assert.Assert(t, err == nil)

	assert.Assert(t, nameFn == filename)
	assert.Assert(t, flagFn == os.O_RDWR|os.O_CREATE|os.O_TRUNC)
	assert.Assert(t, permFn == 0644)
	assert.Assert(t, sizeFn == 0)
	assert.DeepEqual(t, bFn, v)
}
