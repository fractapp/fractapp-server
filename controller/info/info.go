package info

import (
	"encoding/json"
	"fractapp-server/controller"
	"fractapp-server/db"
	"fractapp-server/types"
	"net/http"
)

const (
	TotalRoute = "/total"
)

type Controller struct {
	db            db.DB
	substrateUrls []SubstrateUrl
}

func NewController(db db.DB, substrateUrls map[string]string) *Controller {
	urls := make([]SubstrateUrl, 0)
	for k, v := range substrateUrls {
		urls = append(urls, SubstrateUrl{
			Network: types.ParseNetwork(k),
			Url:     v,
		})
	}
	return &Controller{
		db:            db,
		substrateUrls: urls,
	}
}

func (c *Controller) MainRoute() string {
	return "/info"
}

func (c *Controller) Handler(route string) (func(w http.ResponseWriter, r *http.Request) error, error) {
	switch route {
	case TotalRoute:
		return c.total, nil
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

// info godoc
// @Summary Get total info
// @Description get user by id or blockchain address
// @ID info
// @Tags Info
// @Accept  json
// @Produce json
// @Success 200 {object} TotalInfo
// @Failure 400 {string} string
// @Router /info/total [get]
func (c *Controller) total(w http.ResponseWriter, r *http.Request) error {
	prices := make([]Price, 0)
	for _, v := range types.Currencies {
		price, err := c.db.LastPriceByCurrency(v.String())
		if err != nil && err != db.ErrNoRows {
			return err
		}
		if err == db.ErrNoRows {
			continue
		}

		prices = append(prices, Price{
			Currency: v,
			Value:    price.Price,
		})
	}

	total := &TotalInfo{
		SubstrateUrls: c.substrateUrls,
		Prices:        prices,
	}
	b, err := json.Marshal(&total)
	if err != nil {
		return err
	}

	w.Write(b)
	return nil
}
