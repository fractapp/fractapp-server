package substrate

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"fractapp-server/controller"
	"fractapp-server/db"
	"fractapp-server/types"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	BroadcastRoute = "/broadcast"
	BaseRoute      = "/base"
	TxBaseRoute    = "/txBase"
	FeeRoute       = "/fee"
	BalanceRoute   = "/balance"
)

var (
	InvalidConnectionTxApiErr = errors.New("invalid connection to transaction API")
)

type Controller struct {
	db        db.DB
	txApiHost string
}

func NewController(db db.DB, txApiHost string) *Controller {
	return &Controller{
		db:        db,
		txApiHost: txApiHost,
	}
}

func (c *Controller) MainRoute() string {
	return "/substrate"
}

func (c *Controller) Handler(route string) (func(w http.ResponseWriter, r *http.Request) error, error) {
	switch route {
	case FeeRoute:
		return c.fee, nil
	case TxBaseRoute:
		return c.txBase, nil
	case BaseRoute:
		return c.base, nil
	case BroadcastRoute:
		return c.broadcast, nil
	case BalanceRoute:
		return c.substrateBalance, nil
	}

	return nil, controller.InvalidRouteErr
}
func (c *Controller) ReturnErr(err error, w http.ResponseWriter) {
	switch err {
	case db.ErrNoRows:
		http.Error(w, "", http.StatusNotFound)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

// fee godoc
// @Summary Calculate fee
// @Description calculate fee
// @ID fee
// @Tags Substrate
// @Accept  json
// @Produce json
// @Param sender query string true "sender"
// @Param receiver query string true "receiver"
// @Param value query string true "value"
// @Param currency query int64 true "currency"
// @Success 200 {object} FeeInfo
// @Failure 400 {string} string
// @Router /substrate/fee [get]
func (c *Controller) fee(w http.ResponseWriter, r *http.Request) error {
	tx := r.URL.Query().Get("tx")
	networkInt, err := strconv.ParseInt(r.URL.Query().Get("network"), 10, 32)
	if err != nil {
		return err
	}

	network := types.Network(networkInt)
	resp, err := http.Get(fmt.Sprintf("%s/substrate/fee?network=%s&tx=%s", c.txApiHost, network.String(), tx))
	if err != nil {
		return InvalidConnectionTxApiErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return InvalidConnectionTxApiErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	feeInfo := new(FeeInfo)
	err = json.Unmarshal(body, &feeInfo)
	if err != nil {
		return err
	}

	rsByte, err := json.Marshal(feeInfo)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}

// txBase godoc
// @Summary Get tx base
// @Description calculate fee
// @ID txBase
// @Tags Substrate
// @Accept  json
// @Produce json
// @Param sender query string true "sender"
// @Param sender currency Currency true "currency"
// @Success 200 {object} TxBase
// @Failure 400 {string} string
// @Router /substrate/base [get]
func (c *Controller) txBase(w http.ResponseWriter, r *http.Request) error {
	sender := r.URL.Query().Get("sender")
	networkInt, err := strconv.ParseInt(r.URL.Query().Get("network"), 10, 32)
	if err != nil {
		return err
	}

	network := types.Network(networkInt)
	resp, err := http.Get(fmt.Sprintf("%s/substrate/txBase/%s?network=%s", c.txApiHost, sender, network.String()))
	if err != nil {
		return InvalidConnectionTxApiErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return InvalidConnectionTxApiErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	balance := new(TxBase)
	err = json.Unmarshal(body, &balance)
	if err != nil {
		return err
	}

	rsByte, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}

// base godoc
// @Summary Get substrate base
// @Description substrate base
// @ID base
// @Tags Substrate
// @Accept  json
// @Produce json
// @Param sender currency Currency true "currency"
// @Success 200 {object} TxBase
// @Failure 400 {string} string
// @Router /substrate/base [get]
func (c *Controller) base(w http.ResponseWriter, r *http.Request) error {
	networkInt, err := strconv.ParseInt(r.URL.Query().Get("network"), 10, 32)
	if err != nil {
		return err
	}

	network := types.Network(networkInt)
	resp, err := http.Get(fmt.Sprintf("%s/substrate/base?network=%s", c.txApiHost, network.String()))
	if err != nil {
		return InvalidConnectionTxApiErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return InvalidConnectionTxApiErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	balance := new(Base)
	err = json.Unmarshal(body, &balance)
	if err != nil {
		return err
	}

	rsByte, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}

// broadcast godoc
// @Summary broadcast transaction
// @Description broadcast transaction
// @ID broadcast
// @Tags Substrate
// @Accept  json
// @Produce json
// @Param tx query string true "tx"
// @Param currency query int64 true "currency"
// @Success 200 {object} BroadcastResult
// @Failure 400 {string} string
// @Router /substrate/broadcast [post]
func (c *Controller) broadcast(w http.ResponseWriter, r *http.Request) error {
	tx := r.URL.Query().Get("tx")
	networkInt, err := strconv.ParseInt(r.URL.Query().Get("network"), 10, 32)
	if err != nil {
		return err
	}

	network := types.Network(networkInt)
	resp, err := http.Post(
		fmt.Sprintf("%s/substrate/broadcast?tx=%s&network=%s", c.txApiHost, tx, network.String()),
		"application/json",
		bytes.NewBuffer([]byte{}),
	)
	if err != nil {
		return InvalidConnectionTxApiErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return InvalidConnectionTxApiErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	balance := new(BroadcastResult)
	err = json.Unmarshal(body, &balance)
	if err != nil {
		return err
	}

	rsByte, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}

	return nil
}

// substrateBalance godoc
// @Summary Get substrateBalance by address
// @ID getBalance
// @Tags Substrate
// @Accept  json
// @Produce json
// @Param address query string true "address"
// @Param currency query int true "currency"
// @Success 200 {object} Balance
// @Failure 400 {string} string
// @Router /profile/substrate/balance [get]
func (c *Controller) substrateBalance(w http.ResponseWriter, r *http.Request) error {
	address := r.URL.Query().Get("address")
	currencyInt, err := strconv.ParseInt(r.URL.Query().Get("currency"), 10, 32)
	if err != nil {
		return err
	}
	currency := types.Currency(currencyInt)

	resp, err := http.Get(fmt.Sprintf("%s/substrate/balance/%s?currency=%s", c.txApiHost, address, currency.String()))
	if err != nil {
		return InvalidConnectionTxApiErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return InvalidConnectionTxApiErr
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	balance := new(Balance)
	err = json.Unmarshal(body, &balance)
	if err != nil {
		return err
	}

	rsByte, err := json.Marshal(balance)
	if err != nil {
		return err
	}

	_, err = w.Write(rsByte)
	if err != nil {
		return err
	}
	return nil
}
