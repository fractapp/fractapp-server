package main

import (
	"fmt"

	"github.com/go-pg/migrations/v8"
)

func init() {
	migrations.MustRegisterTx(func(db migrations.DB) error {
		fmt.Println("create table messages...")

		_, err := db.Exec(`create table messages
		(
				id varchar(64) not null
		constraint messages_pk
			primary key,
		value varchar(4096) not null,
		sender_id varchar(64) not null
        		constraint messages_profiles_id_fk_sender
			references profiles,
		receiver_id varchar(64) not null
        		constraint messages_profiles_id_fk_receiver
			references profiles,
		timestamp bigint not null,
		is_delivered bool default false not null
		);`)
		return err
	}, func(db migrations.DB) error {
		fmt.Println("dropping table messages...")

		_, err := db.Exec(`DROP TABLE messages`)
		return err
	})
}
