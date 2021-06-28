package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Contact struct {
	Id          string `bson:"_id"`
	ProfileId   ID     `bson:"profile"`
	PhoneNumber string `pg:"phone_number"`
}

func (db *MongoDB) AllContacts(profileId ID) ([]Contact, error) {
	collection := db.collections[ContactsDB]

	var c []Contact
	res, err := collection.Find(db.ctx, bson.D{
		{"profile", profileId},
	})
	if err != nil {
		return nil, err
	}

	err = res.Decode(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (db *MongoDB) AllMatchContacts(id ID) ([]Profile, error) {
	profile, err := db.ProfileById(id)
	if err != nil {
		return nil, err
	}

	collection := db.collections[ContactsDB]

	var contactsWhoHaveUser []Contact
	res, err := collection.Find(db.ctx, bson.D{
		{"phone_number", profile.PhoneNumber},
	})
	if err != nil {
		return nil, err
	}

	err = res.Decode(contactsWhoHaveUser)
	if err != nil {
		return nil, err
	}

	var usersContacts []Contact
	res, err = collection.Find(db.ctx, bson.D{
		{"profile", id},
	})
	if err != nil {
		return nil, err
	}

	err = res.Decode(usersContacts)
	if err != nil {
		return nil, err
	}

	contacts := make([]Profile, 0)
	usersContactsMap := make(map[ID]string)
	for _, v := range usersContacts {
		usersContactsMap[v.ProfileId] = v.PhoneNumber
	}

	for _, v := range contactsWhoHaveUser {
		if v.ProfileId == id {
			continue
		}

		if _, ok := usersContactsMap[v.ProfileId]; !ok {
			continue
		}

		profile, err := db.ProfileById(v.ProfileId)
		if err != nil {
			continue
		}

		contacts = append(contacts, *profile)
	}

	return contacts, nil
}
