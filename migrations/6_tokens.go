package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table tokens...")

		_, err := db.Exec(`create table tokens
		(
			id varchar(64)
				constraint tokens_profiles_id_fk
			references profiles,
			token varchar not null
		);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table tokens...")

		_, err := db.Exec(`DROP TABLE tokens`)
		return err
	})
}
