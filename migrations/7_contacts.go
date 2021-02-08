package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table contacts...")

		_, err := db.Exec(`create table contacts
		(
			"id"    varchar(64) not null
        		constraint profiles_id_fk
			references profiles,
			"phone_number" varchar(15) not null
		);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table contacts...")

		_, err := db.Exec(`DROP TABLE contacts`)
		return err
	})
}
