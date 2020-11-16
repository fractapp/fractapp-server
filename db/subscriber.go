package db

import "fractapp-server/types"

type Subscriber struct {
	Address string
	Token   string
	Network types.Network `pg:",use_zero"`
}
