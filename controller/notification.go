package controller

import (
	"encoding/json"
	"errors"
	"fractapp-server/db"
	"fractapp-server/types"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/ChainSafe/go-schnorrkel"

	"github.com/go-pg/pg/v10"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	MaxAddressesForToken = 10
	SignMsg              = "It is my firebase token for fractapp:"
	SignTimeout          = 1 * time.Hour
)

var (
	InvalidSignTimeErr        = errors.New("invalid sign time")
	InvalidSignErr            = errors.New("invalid sign")
	InvalidAddressErr         = errors.New("address not equals pubkey")
	MaxAddressCountByTokenErr = errors.New("token limit for addresses exceeded")
)

type NotificationController struct {
	db *pg.DB
}

type UpdateTokenRq struct {
	PubKey  string
	Address string
	Network types.Network
	Sign    string
	Token   string
	Time    int64
}

func NewNotificationController(db *pg.DB) *NotificationController {
	return &NotificationController{
		db: db,
	}
}

func (controller *NotificationController) Subscribe(w http.ResponseWriter, r *http.Request) {
	if err := controller.subscribe(r); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Http error: %s \n", err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (controller *NotificationController) subscribe(r *http.Request) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	log.Printf("Rq body: %s\n", string(b))

	updateTokenRq := UpdateTokenRq{}
	err = json.Unmarshal(b, &updateTokenRq)
	if err != nil {
		return err
	}

	pubKeyBytes, err := hexutil.Decode(updateTokenRq.PubKey)
	if err != nil {
		return err
	}
	pubKey := [32]byte{}
	copy(pubKey[:], pubKeyBytes)

	signBytes, err := hexutil.Decode(updateTokenRq.Sign)
	if err != nil {
		return err
	}

	rqTime := time.Unix(updateTokenRq.Time, 0)
	if rqTime.After(time.Now().Add(SignTimeout)) {
		return InvalidSignTimeErr
	}

	msg := SignMsg + updateTokenRq.Token + strconv.FormatInt(rqTime.Unix(), 10)

	publicKey := &(schnorrkel.PublicKey{})
	err = publicKey.Decode(pubKey)
	if err != nil {
		return err
	}

	// verify sign
	signingContext := schnorrkel.NewSigningContext([]byte("substrate"), []byte(msg))

	sign := [64]byte{}
	copy(sign[:], signBytes)

	signature := &(schnorrkel.Signature{})
	err = signature.Decode(sign)
	if err != nil {
		return err
	}

	if !publicKey.Verify(signature, signingContext) {
		return InvalidSignErr
	}

	address := updateTokenRq.Network.Address(pubKey[:])

	if address != updateTokenRq.Address {
		return InvalidAddressErr
	}

	addressCountByToken, err := controller.db.Model(&db.Subscriber{}).Where("token = ?", updateTokenRq.Token).Count()
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	if addressCountByToken >= MaxAddressesForToken {
		return MaxAddressCountByTokenErr
	}

	err = controller.db.Model(&db.Subscriber{}).Where("address = ?", address).Select()
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	subscriber := &db.Subscriber{
		Address: address,
		Token:   updateTokenRq.Token,
		Network: updateTokenRq.Network,
	}

	if err == pg.ErrNoRows {
		_, err = controller.db.Model(subscriber).Insert()
	} else {
		_, err = controller.db.Model(subscriber).
			Column("token").
			Where("address = ?", subscriber.Address).
			Update()
	}

	if err != nil {
		return err
	}

	return nil
}
