package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table subscribers...")

		_, err := db.Exec(`create table subscribers
(
    "address"    varchar(50) not null
        constraint address_pk
            primary key,
    "token"      varchar,
    "network"   integer
);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table subscribers...")

		_, err := db.Exec(`DROP TABLE subscribers`)
		return err
	})
}
