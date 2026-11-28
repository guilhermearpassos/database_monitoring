package telemetry

import (
	"time"

	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
)

func OpenInstrumentedDB(driver string, connString string) (*sqlx.DB, error) {
	db, err := otelsql.Open(driver, connString)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	return sqlx.NewDb(db, driver), nil
}
