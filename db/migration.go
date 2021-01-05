package db

import "fractapp-server/types"

type AccountMigration struct {
	Number      int64 `pg:",pk"`
	IdFrom      string
	IdTo        string
	Timestamp   int64
	Value       string
	AccountType types.MigrationType
	IsFinished  bool
}
