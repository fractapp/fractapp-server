package db

import (
	"go.mongodb.org/mongo-driver/bson"
)

type Contact struct {
	Id          ID     `bson:"_id"`
	ProfileId   ID     `bson:"profile"`
	PhoneNumber string `pg:"phone_number"`
}

func (db *MongoDB) AllContacts(profileId ID) ([]Contact, error) {
	collection := db.collections[ContactsDB]

	c := make([]Contact, 0)
	res, err := collection.Find(db.ctx, bson.D{
		{"profile", profileId},
	})
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &c)
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

	contactsWhoHaveUser := make([]Contact, 0)
	res, err := collection.Find(db.ctx, bson.D{
		{"phone_number", profile.PhoneNumber},
	})
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &contactsWhoHaveUser)
	if err != nil {
		return nil, err
	}

	usersContacts := make([]Contact, 0)
	res, err = collection.Find(db.ctx, bson.D{
		{"profile", id},
	})
	if err != nil {
		return nil, err
	}

	err = res.All(db.ctx, &usersContacts)
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
