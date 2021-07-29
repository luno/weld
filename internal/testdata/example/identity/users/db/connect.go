package db

import (
	"database/sql"
	"flag"
)

var (
	dbURI = flag.String("users_db", "mysql://root@unix/users?",
		"users DB URI")
	rdbURI = flag.String("users_db_readonly", "", "users DB read replica URI")
)

type UsersDB struct {
	DB        *sql.DB
	ReplicaDB *sql.DB
}

// ReplicaOrMaster returns the replica DB if available, otherwise the master.
func (db *UsersDB) ReplicaOrMaster() *sql.DB {
	if db.ReplicaDB != nil {
		return db.ReplicaDB
	}
	return db.DB
}

func Connect() (*UsersDB, error) {

	return &UsersDB{}, nil
}
