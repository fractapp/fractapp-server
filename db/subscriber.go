package db

import "fractapp-server/types"

type Subscriber struct {
	Address string        `pg:",pk"`
	Token   string        `pg:",use_zero"`
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
func (db *PgDB) SubscribersByRange(from int, limit int) ([]Subscriber, error) {
	sub := make([]Subscriber, 0)
	err := db.Model(&sub).Offset(from).Limit(limit).Select()
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (db *PgDB) SubscribersCount() (int, error) {
	count, err := db.Model(&Subscriber{}).Count()
	if err != nil {
		return 0, err
	}

	return count, nil
}
