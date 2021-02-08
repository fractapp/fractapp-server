package db

type Token struct {
	Id    string `pg:",use_zero"`
	Token string `pg:",use_zero"`
}

func (db *PgDB) IdByToken(token string) (string, error) {
	tokenDb := &Token{}
	err := db.Model(tokenDb).Where("token = ?", token).Select()
	if err != nil {
		return "", err
	}

	return tokenDb.Id, nil
}
func (db *PgDB) TokenById(id string) (string, error) {
	tokenDb := &Token{}
	err := db.Model(tokenDb).Where("id = ?", id).Select()
	if err != nil {
		return "", err
	}

	return tokenDb.Token, nil
}
