package game

import (
	"database/sql"
	"fmt"
	"time"

	// import sqlite package for use with the sql interface
	_ "github.com/mattn/go-sqlite3"
)

const gameTabelCreation = `
	CREATE TABLE IF NOT EXISTS games (
		id text PRIMARY KEY,
		player_white_id text,
		player_black_id text,
		last_moved datetime,
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

const takebackTableCreation = `
	CREATE TABLE IF NOT EXISTS takebacks (
		game_id text PRIMARY KEY,
		fen_snapshot text NOT NULL
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
	db, err := sql.Open("sqlite3", fmt.Sprintf("%v?parseTime=1", path))
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(gameTabelCreation); err != nil {
		return nil, err
	}
	if _, err = db.Exec(challengeTableCreation); err != nil {
		return nil, err
	}
	if _, err = db.Exec(takebackTableCreation); err != nil {
		return nil, err
	}
	store.db = db
	return &store, nil
}

// StoreGame stores a game by ID.
// If a game is already established, only the PGN log is updated
func (s *SqliteStore) StoreGame(ID string, gm *Game) error {
	if _, err := s.RetrieveGame(ID); err == nil {
		stmt, _ := s.db.Prepare("update games set pgn = ?, last_moved = ? where id = ?")
		defer stmt.Close()
		_, err := stmt.Exec(gm.PGN(), gm.LastMoved(), ID)
		return err
	}
	stmt, _ := s.db.Prepare("insert into games (id, player_white_id, player_black_id, last_moved, pgn) values (?, ?, ?, ?, ?)")
	defer stmt.Close()
	_, err := stmt.Exec(ID, gm.Players[White].ID, gm.Players[Black].ID, gm.LastMoved(), gm.PGN())
	return err
}

// RetrieveGame retrieves a game by ID
func (s *SqliteStore) RetrieveGame(ID string) (*Game, error) {
	stmt, err := s.db.Prepare("select player_white_id, player_black_id, last_moved, pgn from games where id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var player1, player2, pgn string
	var lastMoved time.Time
	row := stmt.QueryRow(ID)
	err = row.Scan(&player1, &player2, &lastMoved, &pgn)
	if err != nil {
		return nil, err
	}
	gm, err := NewGameFromPGN(ID, pgn, Player{
		ID: player1,
	}, Player{
		ID: player2,
	})
	if err == nil {
		gm.lastMoved = lastMoved
	}

	return gm, err
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

// StoreTakeback stores a takeback request
// Note: This will overwrite a takeback request if a previous request is left open
func (s *SqliteStore) StoreTakeback(takeback *Takeback) error {
	stmt, _ := s.db.Prepare(`
	insert into takebacks (game_id, fen_snapshot) values (?, ?)
	on conflict (game_id) do update set
		fen_snapshot = ?
	`)
	defer stmt.Close()
	_, err := stmt.Exec(
		takeback.CurrentGame.ID,
		takeback.CurrentGame.FEN(),
		takeback.CurrentGame.FEN(),
	)
	return err
}

// RetrieveTakeback finds a takeback request by a game ID
func (s *SqliteStore) RetrieveTakeback(gameID string) (*Takeback, error) {
	game, err := s.RetrieveGame(gameID)
	if err != nil {
		return nil, err
	}
	stmt, _ := s.db.Prepare("select fen_snapshot from takebacks where game_id = ?")
	defer stmt.Close()
	takeback := Takeback{
		CurrentGame: game,
		FENSnapshot: "",
	}
	row := stmt.QueryRow(gameID)
	err = row.Scan(&takeback.FENSnapshot)
	return &takeback, err
}

// RemoveTakeback removes a takeback request from storage
func (s *SqliteStore) RemoveTakeback(takeback *Takeback) error {
	stmt, _ := s.db.Prepare("delete from takebacks where game_id = ?")
	defer stmt.Close()
	_, err := stmt.Exec(takeback.CurrentGame.ID)
	return err
}
