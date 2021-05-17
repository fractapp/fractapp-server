package db

import (
	"context"

	"github.com/go-pg/pg/v10"
)

type Profile struct {
	Id          string `pg:",pk"`
	Name        string
	Username    string
	PhoneNumber string `pg:"phone_number"`
	Email       string
	IsMigratory bool `pg:",use_zero"`
	AvatarExt   string
	LastUpdate  int64 `pg:",use_zero"`
}

func (db *PgDB) ProfileByMatchedPhoneNumber(contactPhoneNumber string, myPhoneNumber string) (*Profile, error) {
	p := &Profile{}
	if err := db.Model(p).Where("phone_number = ?", contactPhoneNumber).Select(); err != nil {
		return nil, err
	}

	c := &Contact{}
	if err := db.Model(&c).Where("id = ?", p.Id).Where("phone_number = ?", myPhoneNumber).Select(); err != nil {
		return nil, err
	}

	return p, nil
}

func (db *PgDB) SearchUsersByUsername(value string, limit int) ([]Profile, error) {
	var p []Profile
	err := db.Model(&p).Where("username LIKE ?", value+"%").Order("username ASC").Limit(limit).Select()
	if err != nil {
		return nil, err
	}

	return p, nil
}
func (db *PgDB) SearchUsersByEmail(value string) (*Profile, error) {
	p := &Profile{}
	err := db.Model(p).Where("lower(email) = ?", value).Select()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (db *PgDB) ProfileById(id string) (*Profile, error) {
	p := &Profile{}
	err := db.Model(p).Where("id = ?", id).Select()
	if err != nil {
		return nil, err
	}

	return p, nil
}
func (db *PgDB) ProfileByUsername(username string) (*Profile, error) {
	p := &Profile{}
	err := db.Model(p).Where("username = ?", username).Select()
	if err != nil {
		return nil, err
	}

	return p, nil
}
func (db *PgDB) ProfileByAddress(address string) (*Profile, error) {
	a := &Address{}
	err := db.Model(a).Where("address = ?", address).Select()
	if err != nil {
		return nil, err
	}

	p := &Profile{}
	err = db.Model(p).Where("id = ?", a.Id).Select()
	if err != nil {
		return nil, err
	}

	return p, nil
}
func (db *PgDB) ProfileByPhoneNumber(phoneNumber string) (*Profile, error) {
	p := &Profile{}
	err := db.Model(p).Where("phone_number = ?", phoneNumber).Select()
	if err != nil {
		return nil, err
	}

	return p, nil
}
func (db *PgDB) ProfileByEmail(email string) (*Profile, error) {
	p := &Profile{}
	err := db.Model(p).Where("email = ?", email).Select()
	if err != nil {
		return nil, err
	}

	return p, nil
}
func (db *PgDB) UsernameIsExist(username string) (bool, error) {
	return db.Model(&Profile{}).Where("username = ?", username).Exists()
}

func (db *PgDB) CreateProfile(ctx context.Context, profile *Profile, addresses []*Address) error {
	if err := db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		if _, err := tx.Model(profile).Insert(); err != nil {
			return err
		}

		for _, v := range addresses {
			if _, err := tx.Model(v).Insert(); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (db *PgDB) ProfilesCount() (int, error) {
	return db.Model(&Profile{}).Count()
}
