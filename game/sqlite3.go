package game

import (
	"database/sql"

	// import sqlite package for use with the sql interface
	_ "github.com/mattn/go-sqlite3"
)

const gameTabelCreation = `
	CREATE TABLE IF NOT EXISTS games (
		id text PRIMARY KEY,
		player_white_id text,
		player_black_id text,
		pgn text
	);
`

const challengeTableCreation = `
	CREATE TABLE IF NOT EXISTS challenges (
		challenger_id text NOT NULL,
		challenged_id text NOT NULL,
		channel_id text NOT NULL,
		game_id text NOT NULL UNIQUE,
		PRIMARY KEY (challenger_id, challenged_id)
	);
`

// SqliteStore is an implementation of GameStorage and ChallengeStorage interfaces that persists using sqlite3
type SqliteStore struct {
	path string
	db   *sql.DB
}

// NewSqliteStore creates (if not exists) the DB file and structure at the path specified
// It implements the GameStorage and ChallengeStorage interface and is intended as a suitable
// perminent storage of games and challenges
func NewSqliteStore(path string) (*SqliteStore, error) {
	store := SqliteStore{
		path: path,
	}
	db, err := sql.Open("sqlite3", store.path)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(gameTabelCreation); err != nil {
		return nil, err
	}
	if _, err = db.Exec(challengeTableCreation); err != nil {
		return nil, err
	}
	store.db = db
	return &store, nil
}

// StoreGame stores a game by ID.
// If a game is already established, only the PGN log is updated
func (s *SqliteStore) StoreGame(ID string, gm *Game) error {
	if _, err := s.RetrieveGame(ID); err == nil {
		stmt, _ := s.db.Prepare("update games set pgn = ? where id = ?")
		defer stmt.Close()
		_, err := stmt.Exec(gm.PGN(), ID)
		return err
	}
	stmt, _ := s.db.Prepare("insert into games (id, player_white_id, player_black_id, pgn) values (?, ?, ?, ?)")
	defer stmt.Close()
	_, err := stmt.Exec(ID, gm.Players[White].ID, gm.Players[Black].ID, gm.PGN())
	return err
}

// RetrieveGame retrieves a game by ID
func (s *SqliteStore) RetrieveGame(ID string) (*Game, error) {
	stmt, err := s.db.Prepare("select player_white_id, player_black_id, pgn from games where id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var player1, player2, pgn string
	row := stmt.QueryRow(ID)
	err = row.Scan(&player1, &player2, &pgn)
	if err != nil {
		return nil, err
	}
	return NewGameFromPGN(pgn, Player{
		ID: player1,
	}, Player{
		ID: player2,
	})
}

// StoreChallenge only supports inserting new challenges. Challenges should not be updated only inserted/removed
func (s *SqliteStore) StoreChallenge(challenge *Challenge) error {
	stmt, _ := s.db.Prepare("insert into challenges (challenger_id, challenged_id, game_id, channel_id) values (?, ?, ?, ?) ")
	defer stmt.Close()
	_, err := stmt.Exec(challenge.ChallengerID, challenge.ChallengedID, challenge.GameID, challenge.ChannelID)
	return err
}

// RetrieveChallenge retrives a challenge by the challenger and challenged ID
func (s *SqliteStore) RetrieveChallenge(challengerID string, challengedID string) (*Challenge, error) {
	stmt, _ := s.db.Prepare("select game_id, channel_id from challenges where challenger_id = ? and challenged_id = ?")
	defer stmt.Close()
	challenge := Challenge{
		ChallengerID: challengerID,
		ChallengedID: challengedID,
	}
	row := stmt.QueryRow(challengerID, challengedID)
	err := row.Scan(&challenge.GameID, &challenge.ChannelID)
	return &challenge, err
}

// RemoveChallenge removes a challenge from the DB
func (s *SqliteStore) RemoveChallenge(challengerID string, challengedID string) error {
	stmt, _ := s.db.Prepare("delete from challenges where challenger_id = ? and challenged_id = ?")
	defer stmt.Close()
	_, err := stmt.Exec(challengerID, challengedID)
	return err
}
