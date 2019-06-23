package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"

	// register postgres driver
	_ "github.com/lib/pq"
)

// DB represents a database with an open connection
type DB struct {
	*sqlx.DB
}

// NewDB initializes the database connection
func (s *server) NewDB() (*sqlx.DB, error) {

	var db *sqlx.DB
	var err error
	var connectionConfig string

	connectionConfig = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", s.config.dbuser, s.config.dbpass, s.config.dbhost, s.config.dbport, s.config.dbname, s.config.dbsslmode)

	// wait for database to become available
	for {
		db, err = sqlx.Open(s.config.dbdriver, connectionConfig)
		if err != nil {
			// errors here are likely when application is starting up, so just log it and continue.
			log.Println(err)
		}

		err = db.Ping()
		// If no error occurs, stop retrying.
		if err == nil {
			break
		}
		log.Println(err)
		log.Println("Waiting for database to become available")
		time.Sleep(10 * time.Second)
	}

	log.Println("Database connection ready.")
	return db, nil
}
