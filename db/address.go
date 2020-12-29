package db

import "fractapp-server/types"

type Address struct {
	Id      string
	Address string
	Network types.Network `pg:",use_zero"`
}
