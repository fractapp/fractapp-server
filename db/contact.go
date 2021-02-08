package db

type Contact struct {
	Id          string
	PhoneNumber string `pg:"phone_number"`
}

func (db *PgDB) AllContacts(id string) ([]Contact, error) {
	var c []Contact
	err := db.Model(&c).Where("id = ?", id).Select()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (db *PgDB) AllMatchContacts(id string, phoneNumber string) ([]Profile, error) {
	p := make([]Profile, 0)

	err := db.Model(&Contact{}).
		ColumnExpr("p.*").
		Join("JOIN profiles p  ON p.phone_number = contact.phone_number").
		Join("JOIN contacts cTwo").JoinOn("cTwo.id = p.id AND cTwo.phone_number = ?", phoneNumber).
		Where("contact.id = ?", id).Select(&p)

	if err != nil {
		return nil, err
	}

	return p, nil
}
