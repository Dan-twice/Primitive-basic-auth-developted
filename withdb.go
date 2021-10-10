package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

var db *sqlx.DB
var err error

func initDB() {
	pgConnString := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGDATABASE"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
	)
	db, err = sqlx.Open("postgres", pgConnString) // init db
	if err != nil {
		log.Println("Open")
		log.Fatal(err)
	}
}

func retryPing() {
	// retry logic
	retries := 5
	for retries > 0 {
		pingErr := db.Ping() // really connect to database
		if pingErr != nil {
			retries -= 1
			fmt.Println("Retries left ", retries)
			log.Println(pingErr)
			time.Sleep(time.Second)
		} else {
			log.Printf("Left Ping")
			break
		}
	}
}

func migrater() {
	migrations := &migrate.FileMigrationSource{
		Dir: "./",
	}

	n, err := migrate.Exec(db.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		fmt.Println("Error occured:", err)
	}
	fmt.Printf("Applied %d migrations!\n", n)
}

func DdDealer() {
	initDB()
	retryPing()
	migrater()
}
