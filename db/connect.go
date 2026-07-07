package db

import (
    "database/sql"
    "log"
    "os"
    "strconv"
    "github.com/lib/pq"
)


func Connect() (*sql.DB, error) {


    port, _ := strconv.ParseUint(os.Getenv("PGPORT"), 10, 16)

    cfg := pq.Config{
		Host:     os.Getenv("PGHOST"),
		Port:     uint16(port),
		User:     os.Getenv("PGUSER"),
		Password: os.Getenv("PGPASSWORD"),
		Database:   os.Getenv("PGDB"),
                SSLMode:    "disable",
	}

    c, err := pq.NewConnectorConfig(cfg)
    if err != nil {
       log.Fatal(err)
    }

    db := sql.OpenDB(c)

    err = db.Ping()
    if err != nil {
	db.Close()
	log.Fatal(err)
    }

    return db, nil
}

func CreateTable () {

    db, err := Connect()
    if err != nil {
       log.Fatal(err)
    }

    _, err = db.Exec("CREATE TABLE IF NOT EXISTS wallets (valletId TEXT PRIMARY KEY, amount INTEGER)")

    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

}

func CheckValletid (valletid string) bool {
    var s sql.NullString
    db, err := Connect()
    if err != nil {
       log.Fatal(err)
    }

    err = db.QueryRow("SELECT valletid FROM wallets WHERE valletid = $1", valletid).Scan(&s)
    if s.Valid { return true } else { return false }
}

