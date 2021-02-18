package db

import (
	"context"
	"fractapp-server/notification"

	"github.com/go-pg/pg/v10"
)

var (
	ErrNoRows = pg.ErrNoRows
)

type DB interface {
	SubscribersCountByToken(token string) (int, error)
	SubscriberByAddress(address string) (*Subscriber, error)
	AuthByValue(value string, codeType notification.NotificatorType, checkType notification.CheckType) (*Auth, error)
	AddressesById(id string) ([]Address, error)
	ProfileById(id string) (*Profile, error)
	ProfileByAddress(address string) (*Profile, error)
	ProfileByUsername(username string) (*Profile, error)
	AddressIsExist(address string) (bool, error)
	UsernameIsExist(username string) (bool, error)
	SearchUsersByUsername(value string, limit int) ([]Profile, error)
	SearchUsersByEmail(value string) (*Profile, error)
	ProfileByPhoneNumber(contactPhoneNumber string, myPhoneNumber string) (*Profile, error)
	CreateProfile(ctx context.Context, profile *Profile, addresses []*Address) error
	IdByToken(token string) (string, error)
	TokenById(id string) (string, error)
	AllContacts(id string) ([]Contact, error)
	AllMatchContacts(id string, phoneNumber string) ([]Profile, error)

	SubscribersByRange(from int, limit int) ([]Subscriber, error)
	SubscribersCount() (int, error)

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
