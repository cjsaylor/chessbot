package integration

import (
	"database/sql"

	// import sqlite package for use with the sql interface
	_ "github.com/mattn/go-sqlite3"
)

const authTableCreation = `
	CREATE TABLE IF NOT EXISTS authorizations (
		id text PRIMARY KEY,
		token text
	);
`

// SqliteStore is an implementation of GameStorage and ChallengeStorage interfaces that persists using sqlite3
type SqliteStore struct {
	path string
	db   *sql.DB
}

// NewSqliteStore creates (if not exists) the DB file and structure at the path specified
// It implements the AuthStorage interface and is intended as a suitable
// perminent storage of oauth tokens
func NewSqliteStore(path string) (*SqliteStore, error) {
	store := SqliteStore{
		path: path,
	}
	db, err := sql.Open("sqlite3", store.path)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(authTableCreation); err != nil {
		return nil, err
	}
	store.db = db
	return &store, nil
}

func (s *SqliteStore) StoreAuthToken(teamID string, oauthToken string) error {
	stmt, _ := s.db.Prepare(`
		insert into authorizations (id, token) values (?, ?)
		on conflict(id) do update set token = ?
	`)
	defer stmt.Close()
	_, err := stmt.Exec(teamID, oauthToken, oauthToken)
	return err
}

func (s *SqliteStore) GetAuthToken(teamID string) (string, error) {
	stmt, _ := s.db.Prepare("select token from authorizations where id = ?")
	defer stmt.Close()
	var token string
	row := stmt.QueryRow(teamID)
	err := row.Scan(&token)
	return token, err
}
