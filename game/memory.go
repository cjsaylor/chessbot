package game

import (
	"fmt"
	"log"
)

type MemoryStore struct {
	games      map[string]*Game
	challenges map[string]*Challenge
}

func NewMemoryStore() *MemoryStore {
	store := MemoryStore{
		games:      make(map[string]*Game, 10),
		challenges: make(map[string]*Challenge, 10),
	}
	return &store
}

func (m *MemoryStore) RetrieveGame(id string) (*Game, error) {
	gm, ok := m.games[id]
	if !ok {
		return nil, fmt.Errorf("Game by %v not found", id)
	}
	return gm, nil
}

func (m *MemoryStore) StoreGame(id string, game *Game) error {
	m.games[id] = game
	return nil
}

func (m *MemoryStore) RetrieveChallenge(challengerID string, challengedID string) (*Challenge, error) {
	challenge, ok := m.challenges[challengerID+challengedID]
	if !ok {
		return nil, fmt.Errorf("Challenge %v%v not found", challengerID, challengedID)
	}
	return challenge, nil
}

func (m *MemoryStore) StoreChallenge(c *Challenge) error {
	key := c.ChallengerID + c.ChallengedID
	m.challenges[key] = c
	log.Printf("[memory store] Saving challenge game %v\n", c.GameID)
	return nil
}
