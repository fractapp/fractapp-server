package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table account_migrations...")

		_, err := db.Exec(`create table account_migrations
(
	"number" bigserial
		constraint account_migrations_pk
	primary key,
	"is_valid" boolean,
	"id_from" varchar(64) not null
		constraint account_migrations_profiles_id_fk_from
	references profiles,
	"id_to" varchar(64) not null
		constraint account_migrations_profiles_id_fk_to
	references profiles,
	"timestamp" bigint not null,
	"value" varchar(320) not null,
	"account_type" int not null,
	"is_finished" boolean not null
);
`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table account_migrations...")

		_, err := db.Exec(`DROP TABLE account_migrations`)
		return err
	})
}
