package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("add is_chat_bot to profile table...")

		_, err := db.Exec(`
alter table profiles ADD is_chat_bot bool default false not null;
		`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping is_chat_bot from profile...")

		_, err := db.Exec(`
alter table profiles DROP column is_chat_bot;
		`)
		return err
	})
}
