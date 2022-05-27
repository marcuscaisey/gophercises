package repo

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/marcuscaisey/gophercises/2-urlshort/v2/errors"
	"github.com/marcuscaisey/gophercises/2-urlshort/v2/errors/codes"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteURLRepository struct {
	db DB
}

type DB interface {
	Exec(string, ...any) (sql.Result, error)
	QueryRow(string, ...any) *sql.Row
}

func NewSQLiteURLRepository(db DB) *SQLiteURLRepository {
	return &SQLiteURLRepository{db: db}
}

func (r *SQLiteURLRepository) Migrate() error {
	const createURLsTableQuery = `
		CREATE TABLE IF NOT EXISTS urls (
			short_path TEXT PRIMARY KEY,
			long_url TEXT NOT NULL
		) WITHOUT ROWID;
	`
	log.Println("Ensuring that urls table exists.")

	_, err := r.db.Exec(createURLsTableQuery)
	if err != nil {
		return fmt.Errorf("create URLs table: %w", err)
	}
	return nil
}

func (r *SQLiteURLRepository) MustMigrate() {
	if err := r.Migrate(); err != nil {
		panic(fmt.Sprintf("migrate: %s", err))
	}
}

func (r *SQLiteURLRepository) Create(shortPath, longURL string) error {
	const insertURLQuery = "INSERT OR IGNORE INTO urls (short_path, long_url) VALUES ($1, $2);"
	result, err := r.db.Exec(insertURLQuery, shortPath, longURL)
	if err != nil {
		return fmt.Errorf("insert (%q, %q) into urls (short_path, long_url): %w", shortPath, longURL, err)
	}
	if rowsAffected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("get rows affected by insert: %w", err)
	} else if rowsAffected == 0 {
		return errors.New(codes.AlreadyExists)
	}
	return nil
}

func (r *SQLiteURLRepository) Get(shortPath string) (string, error) {
	const selectLongURLQuery = "SELECT long_url FROM urls WHERE short_path = $1;"
	var longURL string
	if err := r.db.QueryRow(selectLongURLQuery, shortPath).Scan(&longURL); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New(codes.NotFound)
		}
		return "", fmt.Errorf("select url with short_path = %q: %w", shortPath, err)
	}
	return longURL, nil
}
