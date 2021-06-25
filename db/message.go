package db

type Message struct {
	Id          string `pg:",pk"`
	Value       string `pg:",use_zero"`
	SenderId    string `pg:",use_zero"`
	ReceiverId  string `pg:",use_zero"`
	Timestamp   int64  `pg:",use_zero"`
	IsDelivered bool
}

func (db *PgDB) NotDeliveredMessages() ([]Message, error) {
	var messages []Message
	err := db.Model(&messages).Where("is_delivered = ?", false).Order("timestamp ASC").Select()

	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (db *PgDB) MessagesByReceiver(id string) ([]Message, error) {
	var messages []Message
	err := db.Model(&messages).
		Where("receiver_id = ?", id).Where("is_delivered = ?", false).Order("timestamp ASC").Select()

	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (db *PgDB) MessagesBySenderAndReceiver(sender string, receiver string) ([]Message, error) {
	var messages []Message
	err := db.Model(&messages).
		Where("sender_id = ?", sender).Where("receiver_id = ?", receiver).Where("is_delivered = ?", false).Order("timestamp ASC").Select()

	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (db *PgDB) UpdateDeliveredMessage(id string, receiverId string) error {
	_, err := db.Model(&Message{
		IsDelivered: true,
	}).Where("receiver_id = ?", receiverId).Where("id = ?", id).Column("is_delivered").Update()
	if err != nil {
		return err
	}

	return nil
}
