package db

import "fractapp-server/types"

type Auth struct {
	PhoneNumber string `pg:"phone_number,pk"`
	Code        string
	Attempts    int32
	Count       int32 `pg:",use_zero"`
	Timestamp   int64
	Type        types.CodeType
	CheckType   types.CheckType
}

func (db *PgDB) AuthByPhoneNumber(phoneNumber string) (*Auth, error) {
	auth := &Auth{}
	err := db.Model(auth).Where("phone_number = ?", phoneNumber).Select()
	if err != nil {
		return nil, nil
	}

	return auth, nil
}
