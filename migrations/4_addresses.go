package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table addresses...")

		_, err := db.Exec(`create table addresses
(
	"id"    varchar(64) not null
        constraint profiles_id_fk
			references profiles,
  	"address"    varchar(50) not null
        constraint address_pk
            primary key,
    "network"   integer
);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table addresses...")

		_, err := db.Exec(`DROP TABLE addresses`)
		return err
	})
}
