package db

import (
	"database/sql"
	"flag"
)

var (
	dbURI = flag.String("exchange_db", "mysql://root@unix/exchange?",
		"exchange DB URI")
	rdbURI = flag.String("exchange_db_readonly", "", "exchange DB read replica URI")
)

type ExchangeDB struct {
	DB        *sql.DB
	ReplicaDB *sql.DB
}

// ReplicaOrMaster returns the replica DB if available, otherwise the master.
func (db *ExchangeDB) ReplicaOrMaster() *sql.DB {
	if db.ReplicaDB != nil {
		return db.ReplicaDB
	}
	return db.DB
}

func Connect() (*ExchangeDB, error) {

	return &ExchangeDB{}, nil
}
