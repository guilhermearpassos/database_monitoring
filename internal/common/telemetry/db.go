package telemetry

import (
	"github.com/XSAM/otelsql"
	"github.com/jmoiron/sqlx"
)

func OpenInstrumentedDB(driver string, connString string) (*sqlx.DB, error) {
	newDriverName, err := otelsql.Register(driver)
	if err != nil {
		return nil, err
	}

	return sqlx.Open(newDriverName, connString)
}
