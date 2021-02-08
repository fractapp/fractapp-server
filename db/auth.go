package db

import "fractapp-server/types"

type Auth struct {
	Value     string          `pg:"value,pk"`
	IsValid   bool            `pg:"is_valid,use_zero"`
	Code      string          `pg:",use_zero"`
	Attempts  int32           `pg:",use_zero"`
	Count     int32           `pg:",use_zero"`
	Timestamp int64           `pg:",use_zero"`
	Type      types.CodeType  `pg:",use_zero"`
	CheckType types.CheckType `pg:",use_zero"`
}

func (db *PgDB) AuthByValue(value string, codeType types.CodeType, checkType types.CheckType) (*Auth, error) {
	auth := &Auth{}
	err := db.Model(auth).Where("value = ?", value).
		Where("type = ?", codeType).
		Where("check_type = ?", checkType).Select()
	if err != nil {
		return nil, err
	}

	return auth, nil
}
