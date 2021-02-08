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
    "value"    varchar(256) not null
        constraint value_pk
            primary key,
    "code"      varchar(6) not null,
 	"attempts"   integer,
    "count"   integer,
 	"timestamp"   bigint,
	"type"   integer,
	"check_type"   integer
);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table auths...")

		_, err := db.Exec(`DROP TABLE auths`)
		return err
	})
}
