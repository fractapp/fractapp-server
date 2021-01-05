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
	Twitter     string
	IsMigratory bool
}

func (db *PgDB) ProfileById(id string) (*Profile, error) {
	p := &Profile{}
	err := db.Model(p).Where("id = ?", id).Select()
	if err != nil {
		return nil, nil
	}

	return p, nil
}

func (db *PgDB) UsernameIsExist(username string) (bool, error) {
	return db.Model(&Profile{}).Where("username = ?", username).Exists()
}

func (db *PgDB) CreateProfile(ctx context.Context, profile *Profile, addresses []*Address) error {
	if err := db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		if _, err := tx.Model(profile).Insert(tx); err != nil {
			return err
		}

		for _, v := range addresses {
			if _, err := db.Model(v).Insert(); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
