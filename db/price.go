package db

type Price struct {
	Timestamp int64   `pg:",use_zero"`
	Currency  string  `pg:",use_zero"`
	Price     float32 `pg:",use_zero"`
}

func (db *PgDB) Prices(currency string, startTime int64, endTime int64) ([]Price, error) {
	var price []Price
	err := db.Model(&price).Where("currency = ?", currency).
		Where("timestamp >= ?", startTime).
		Where("timestamp <= ?", endTime).Select()
	if err != nil {
		return nil, err
	}

	return price, nil
}

func (db *PgDB) LastPriceByCurrency(currency string) (*Price, error) {
	price := &Price{}
	err := db.Model(price).Where("currency = ?", currency).Order("timestamp DESC").Limit(1).Select()
	if err != nil {
		return nil, err
	}

	return price, nil
}
