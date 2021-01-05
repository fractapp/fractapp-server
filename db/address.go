package db

import "fractapp-server/types"

type Address struct {
	Id      string
	Address string        `pg:",pk"`
	Network types.Network `pg:",use_zero"`
}

func (db *PgDB) AddressIsExist(address string) (bool, error) {
	return db.Model(&Address{}).Where("address = ?", address).Exists()
}
