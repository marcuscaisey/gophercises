package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/marcuscaisey/gophercises/2-urlshort/v2/repo"
	"github.com/marcuscaisey/gophercises/2-urlshort/v2/server"
)

var sqliteFile = flag.String("db-file", "db.sqlite", "Path to SQLite DB")
var inMemory = flag.Bool("in-memory", false, "Whether to use an in memory DB instead of SQLite")
var port = flag.Uint("port", 8080, "Port to serve on")

func main() {
	flag.Parse()

	var urlServer *server.Server
	if *inMemory {
		log.Println("Using in-memory DB.")
		repo := repo.NewInMemoryURLRepository()
		urlServer = server.New(repo)

	} else {
		log.Printf("Using SQLite DB at %s.", *sqliteFile)
		db := mustOpenSQLiteDB(*sqliteFile)
		repo := repo.NewSQLiteURLRepository(db)
		repo.MustMigrate()
		urlServer = server.New(repo)
	}

	log.Fatal(urlServer.Run(*port))
}

func mustOpenSQLiteDB(path string) *sql.DB {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		panic(fmt.Sprintf("connect to sqlite db at %q: %s", path, err))
	}
	return db
}
