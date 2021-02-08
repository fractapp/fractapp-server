package db

import "fractapp-server/types"

type Address struct {
	Id      string
	Address string        `pg:",pk"`
	Network types.Network `pg:",use_zero"`
}

func (db *PgDB) AddressesById(id string) ([]Address, error) {
	var addresses []Address
	if err := db.Model(&addresses).Where("id = ?", id).Select(); err != nil {
		return nil, err
	}
	return addresses, nil
}

func (db *PgDB) AddressIsExist(address string) (bool, error) {
	return db.Model(&Address{}).Where("address = ?", address).Exists()
}
