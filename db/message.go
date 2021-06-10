package db

type Message struct {
	Id          string `pg:",pk"`
	Value       string `pg:",use_zero"`
	SenderId    string `pg:",use_zero"`
	ReceiverId  string `pg:",use_zero"`
	Timestamp   int64  `pg:",use_zero"`
	IsDelivered bool
}

func (db *PgDB) GetMessagesByReceiver(id string) ([]Message, error) {
	var messages []Message
	err := db.Model(&messages).
		Where("receiver_id = ?", id).Where("is_delivered = ?", false).Order("timestamp ASC").Select()

	if err != nil {
		return nil, err
	}

	return messages, nil
}
