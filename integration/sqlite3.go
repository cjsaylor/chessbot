package integration

import (
	"database/sql"
	"fmt"

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

// NewSqliteStoreWithDB creates the auth table structure using an existing *sql.DB connection.
// This avoids opening a second connection pool to the same SQLite file.
func NewSqliteStoreWithDB(db *sql.DB) (*SqliteStore, error) {
	if _, err := db.Exec(authTableCreation); err != nil {
		return nil, err
	}
	return &SqliteStore{db: db}, nil
}

// StoreAuthToken stores the oauth token granted by slack user for a given team ID
func (s *SqliteStore) StoreAuthToken(teamID string, oauthToken string) error {
	stmt, _ := s.db.Prepare(`
		insert into authorizations (id, token) values (?, ?)
		on conflict(id) do update set token = ?
	`)
	defer stmt.Close()
	_, err := stmt.Exec(teamID, oauthToken, oauthToken)
	return err
}

// GetAuthToken retrieves an oauth token for use with slack given a team ID
func (s *SqliteStore) GetAuthToken(teamID string) (string, error) {
	debugf("GetAuthToken: querying db at path=%s for teamID=%q", s.path, teamID)
	stmt, err := s.db.Prepare("select token from authorizations where id = ?")
	if err != nil {
		return "", fmt.Errorf("GetAuthToken prepare error: %w", err)
	}
	defer stmt.Close()
	var token string
	row := stmt.QueryRow(teamID)
	err = row.Scan(&token)
	if err != nil {
		// Log what's actually in the table for diagnosis
		rows, qErr := s.db.Query("select id from authorizations")
		if qErr == nil {
			defer rows.Close()
			var ids []string
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					ids = append(ids, id)
				}
			}
			debugf("GetAuthToken: lookup failed for teamID=%q, existing IDs=%v, error=%v", teamID, ids, err)
		}
		return "", err
	}
	return token, nil
}
