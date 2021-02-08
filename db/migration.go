package db

import "fractapp-server/types"

type AccountMigration struct {
	Number      int64               `pg:",pk"`
	IdFrom      string              `pg:",use_zero"`
	IdTo        string              `pg:",use_zero"`
	Timestamp   int64               `pg:",use_zero"`
	Value       string              `pg:",use_zero"`
	AccountType types.MigrationType `pg:",use_zero"`
	IsFinished  bool                `pg:",use_zero"`
}
