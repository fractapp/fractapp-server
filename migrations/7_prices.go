package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table prices...")

		_, err := db.Exec(`create table prices
		(
			"timestamp" bigint not null,
			"price" real not null,
			"currency" varchar(5)
		);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table prices...")

		_, err := db.Exec(`DROP TABLE prices`)
		return err
	})
}
