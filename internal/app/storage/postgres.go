package storage

import (
	"database/sql"
	"fmt"
)

// Open will open a SQL connection with the provided data source name.
// Callers of Open need to ensure the connection is eventually closed via the
// db.Close() method.
func Open(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("can't init db: %w", err)
	}
	return db, nil
}
