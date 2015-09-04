package store

import (
	"database/sql"
	"flag"
	"log"

	_ "github.com/lib/pq"
	"github.com/bbhmakerlab/debug"
)

const (
	createEntrySQL = `
	id serial PRIMARY KEY,
	name text NOT NULL,
	address1 text NOT NULL,
	address2 text NOT NULL,
	city text NOT NULL,
	state text NOT NULL,
	country text NOT NULL,
	postal_code text NOT NULL,
	email text NOT NULL,
	updated_at timestamp NOT NULL,
	created_at timestamp NOT NULL,
	CONSTRAINT u_constraint UNIQUE (email)`
)

var db *sql.DB
var dataSource = flag.String("datasource", "user=bbh dbname=ikea-catalogue-giveaway-2016 sslmode=disable password=Lion@123", "SQL data source")

func Init() {
	var err error

	db, err = sql.Open("postgres", *dataSource)
	if err != nil {
		debug.Fatal(err)
	}

	create := func(name, content string) {
		if err != nil {
			debug.Fatal(err)
		}
		err = createTable(name, content)
	}

	create("entry", createEntrySQL)
}

func createTable(name, content string) error {
	if exists, err := tableExists(name); err != nil {
		return err
	} else if exists {
		return nil
	}

	if _, err := db.Exec("CREATE TABLE " + name + "(" + content + ")"); err != nil {
		debug.Warn(err)
		return err
	}

	log.Println("created table:", name)
	return nil
}

func tableExists(name string) (bool, error) {
	var q = `SELECT * from information_schema.tables WHERE table_schema = 'public' AND table_name = '` + name + `'`

	rows, err := db.Query(q)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}

func InsertEntry(name, address1, address2, city, state, country, postalCode, email string) error {
	const rawSQL = `
	INSERT INTO entry (name, address1, address2, city, state, country, postal_code, email, updated_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())`

	_, err := db.Exec(rawSQL, name, address1, address2, city, state, country, postalCode, email)
	return err
}

func CountEntries() (int64, error) {
	const rawSQL = `SELECT COUNT(*) FROM entry`

	var count int64
	if err := db.QueryRow(rawSQL, ).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}
