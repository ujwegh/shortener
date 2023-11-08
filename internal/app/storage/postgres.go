package storage

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(dataSourceName string) *sql.DB {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}
