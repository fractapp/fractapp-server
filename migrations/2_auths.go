package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table auths...")

		_, err := db.Exec(`create table auths
(
    "phone_number"    varchar(15) not null
        constraint phone_number_pk
            primary key,
    "code"      varchar(6) not null,
    "count"   integer,
 	"timestamp"   integer
);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table auths...")

		_, err := db.Exec(`DROP TABLE auths`)
		return err
	})
}
