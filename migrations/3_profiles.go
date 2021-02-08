package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table profiles...")

		_, err := db.Exec(`create table profiles
(
	"id"    varchar(64) not null
        constraint profile_id_pk
            primary key,
    "phone_number" varchar(15)
        constraint profile_phone_number_pk
            unique,
    "name"	varchar(32),
	"username" varchar(32)
		constraint username_pk
            unique,
    "email"   varchar(256)
		constraint email_pk
            unique,
	"is_migratory" boolean,
	"avatar_ext" varchar(4),
	"last_update"  bigint not null
);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table profiles...")

		_, err := db.Exec(`DROP TABLE profiles`)
		return err
	})
}
