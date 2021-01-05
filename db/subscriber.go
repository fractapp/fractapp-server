package db

import "fractapp-server/types"

type Subscriber struct {
	Address string `pg:",pk"`
	Token   string
	Network types.Network `pg:",use_zero"`
}

func (db *PgDB) SubscribersCountByToken(token string) (int, error) {
	return db.Model(&Subscriber{}).
		Where("token = ?", token).Count()
}
func (db *PgDB) SubscriberByAddress(address string) (*Subscriber, error) {
	sub := &Subscriber{}
	err := db.Model(sub).Where("address = ?", address).Select()
	if err != nil {
		return nil, err
	}

	return sub, nil
}
