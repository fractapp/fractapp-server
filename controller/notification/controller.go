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
)

const (
	MaxAddressesForToken = 10
	SubscribeMsg         = "It is my firebase token for fractapp:"
	SubscribeRoute       = "/subscribe"
)

var (
	MaxAddressCountByTokenErr = errors.New("token limit for addresses exceeded")
)

type Controller struct {
	db db.DB
}

func NewController(db db.DB) *Controller {
	return &Controller{
		db: db,
	}
}

func (c *Controller) MainRoute() string {
	return "/notification"
}
func (c *Controller) Handler(route string) (func(r *http.Request) error, error) {
	switch route {
	case SubscribeRoute:
		return c.subscribe, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	http.Error(w, err.Error(), http.StatusBadRequest)
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

	subsCountByToken, err := c.db.SubscribersCountByToken(updateTokenRq.Token)
	if err != nil && err != db.ErrNoRows {
		return err
	}

	if subsCountByToken >= MaxAddressesForToken {
		return MaxAddressCountByTokenErr
	}

	sub, err := c.db.SubscriberByAddress(updateTokenRq.Address)
	if err != nil && err != db.ErrNoRows {
		return err
	}

	if err == db.ErrNoRows {
		sub = &db.Subscriber{
			Address: updateTokenRq.Address,
			Token:   updateTokenRq.Token,
			Network: updateTokenRq.Network,
		}
	} else {
		sub.Token = updateTokenRq.Token
	}

	if err == db.ErrNoRows {
		err = c.db.Insert(sub)
		if err != nil {
			return err
		}
	} else {
		err = c.db.UpdateByPK(sub)
		if err != nil {
			return err
		}
	}

	return nil
}
