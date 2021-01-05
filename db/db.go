package db

import (
	"context"

	"github.com/go-pg/pg/v10"
)

var (
	ErrNoRows = pg.ErrNoRows
)

type DB interface {
	SubscribersCountByToken(token string) (int, error)
	SubscriberByAddress(address string) (*Subscriber, error)
	AuthByPhoneNumber(phoneNumber string) (*Auth, error)
	ProfileById(id string) (*Profile, error)
	AddressIsExist(address string) (bool, error)
	UsernameIsExist(username string) (bool, error)
	CreateProfile(ctx context.Context, profile *Profile, addresses []*Address) error

	Insert(value interface{}) error
	UpdateByPK(value interface{}) error
	Update(value interface{}, condition string, params ...interface{}) error
}

type PgDB pg.DB

func (db *PgDB) Insert(value interface{}) error {
	_, err := db.Model(value).Insert()
	if err != nil {
		return err
	}

	return nil
}
func (db *PgDB) UpdateByPK(value interface{}) error {
	_, err := db.Model(value).WherePK().Update()
	if err != nil {
		return err
	}

	return nil
}
func (db *PgDB) Update(value interface{}, condition string, params ...interface{}) error {
	_, err := db.Model(value).Where(condition, params).Update()
	if err != nil {
		return err
	}

	return nil
}
