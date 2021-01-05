package controller

import (
	"log"
	"net/http"
)

type Controller interface {
	MainRoute() string
	Handler(route string) (func(r *http.Request) error, error)
	ReturnErr(err error, w http.ResponseWriter)
}

func Route(c Controller, route string) func(w http.ResponseWriter, r *http.Request) {
	h, err := c.Handler(route)
	if err != nil {
		log.Printf("Route error: %s \n", err.Error())
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(r); err != nil {
			log.Printf("Http error: %s \n", err.Error())

			c.ReturnErr(err, w)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
