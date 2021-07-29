package db

import (
	"database/sql"
	"flag"
)

var (
	dbURI = flag.String("email_db", "mysql://root@unix/email?",
		"email DB URI")
	rdbURI = flag.String("email_db_readonly", "", "email DB read replica URI")
)

type EmailDB struct {
	DB        *sql.DB
	ReplicaDB *sql.DB
}

// ReplicaOrMaster returns the replica DB if available, otherwise the master.
func (db *EmailDB) ReplicaOrMaster() *sql.DB {
	if db.ReplicaDB != nil {
		return db.ReplicaDB
	}
	return db.DB
}

func Connect() (*EmailDB, error) {

	return &EmailDB{}, nil
}
