package notification

import (
	"encoding/json"
	"errors"
	"fractapp-server/controller"
	"fractapp-server/db"
	"fractapp-server/utils"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
)

const (
	MaxAddressesForToken = 10
	SubscribeMsg         = "It is my firebase token for fractapp:"
)

var (
	MaxAddressCountByTokenErr = errors.New("token limit for addresses exceeded")
)

type Controller struct {
	db *pg.DB
}

func NewController(db *pg.DB) *Controller {
	return &Controller{
		db: db,
	}
}

func (c *Controller) Subscribe(w http.ResponseWriter, r *http.Request) {
	if err := c.subscribe(r); err != nil {
		log.Printf("Http error: %s \n", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) subscribe(r *http.Request) error {
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

	rqTime := time.Unix(updateTokenRq.Timestamp, 0)
	if rqTime.After(time.Now().Add(controller.SignTimeout)) {
		return controller.InvalidSignTimeErr
	}

	msg := SubscribeMsg + updateTokenRq.Token + strconv.FormatInt(rqTime.Unix(), 10)
	pubKey, err := utils.ParsePubKey(updateTokenRq.PubKey)
	if err != nil {
		return err
	}

	if updateTokenRq.Network.Address(pubKey[:]) != updateTokenRq.Address {
		return controller.InvalidAddressErr
	}

	if err := utils.Verify(pubKey, msg, updateTokenRq.Sign); err != nil {
		return err
	}

	addressCountByToken, err := c.db.Model(&db.Subscriber{}).
		Where("token = ?", updateTokenRq.Token).Count()
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	if addressCountByToken >= MaxAddressesForToken {
		return MaxAddressCountByTokenErr
	}

	err = c.db.Model(&db.Subscriber{}).Where("address = ?", updateTokenRq.Address).Select()
	if err != nil && err != pg.ErrNoRows {
		return err
	}

	subscriber := &db.Subscriber{
		Address: updateTokenRq.Address,
		Token:   updateTokenRq.Token,
		Network: updateTokenRq.Network,
	}
	if err == pg.ErrNoRows {
		_, err = c.db.Model(subscriber).Insert()
		if err != nil {
			return err
		}
	} else {
		_, err = c.db.Model(subscriber).
			Column("token").
			Where("address = ?", subscriber.Address).
			Update()
		if err != nil {
			return err
		}
	}

	return nil
}
