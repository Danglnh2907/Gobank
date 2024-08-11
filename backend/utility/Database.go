package utility

import (
	//Import standard library
	"database/sql"
	"fmt"
	"time"

	//Import 3rd party package
	_ "github.com/lib/pq"
)

var db *sql.DB

func ConnectDB(dbname string) (*sql.DB, error) {
	const (
		host     = "localhost"
		user     = "postgres"
		password = "test"
		port     = 5432
	)

	var err error

	connectStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err = sql.Open("postgres", connectStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(5 * time.Minute)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func GetDB() *sql.DB {
	return db
}

func InitializeTable() error {
	//Create TABLE users
	sqlQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(10) PRIMARY KEY,
			email VARCHAR(30),
			password VARCHAR(64),
			fullname VARCHAR(30),
			balance DECIMAL,
			exp INT,
			state VARCHAR(10)
		)
	`
	_, err := db.Exec(sqlQuery)
	if err != nil {
		return err
	}

	//Create TABLE admins
	sqlQuery = `
		CREATE TABLE IF NOT EXISTS admins (
			id SERIAL PRIMARY KEY,
			email VARCHAR(30),
			password VARCHAR(64),
			fullname VARCHAR(30)
		)
	`
	_, err = db.Exec(sqlQuery)
	if err != nil {
		return err
	}

	//Create TABLE transactions
	sqlQuery = `
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			date DATE,
			debit VARCHAR(10),
			credit VARCHAR(10),
			beneficiary VARCHAR(50),
			amount DECIMAL,
			description VARCHAR(255)
		)
	`
	_, err = db.Exec(sqlQuery)
	if err != nil {
		return err
	}

	return nil
}
