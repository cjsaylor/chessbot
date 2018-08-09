package game

import (
	"fmt"
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

func (m *MemoryStore) RetrieveGame(ID string) (*Game, error) {
	gm, ok := m.games[ID]
	if !ok {
		return nil, fmt.Errorf("Game by %v not found", ID)
	}
	return gm, nil
}

func (m *MemoryStore) StoreGame(ID string, game *Game) error {
	m.games[ID] = game
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
	return nil
}

func (m *MemoryStore) RemoveChallenge(challengerID string, challengedID string) error {
	key := challengerID + challengedID
	delete(m.challenges, key)
	return nil
}
