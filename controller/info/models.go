package info

import "fractapp-server/types"

type SubstrateUrl struct {
	Network types.Network `json:"network"`
	Url     string        `json:"url"`
}
type Price struct {
	Currency types.Currency `json:"currency"`
	Value    float32        `json:"value"`
}
type TotalInfo struct {
	Prices []Price `json:"prices"`
}
