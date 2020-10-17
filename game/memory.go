package game

import (
	"fmt"
)

// MemoryStore implements the Game and Challenge storage interfaces and holds all state in memory
// Once the MemoryStore instance is released, all data in that storage is lost
type MemoryStore struct {
	games      map[string]*Game
	challenges map[string]*Challenge
	takebacks  map[string]*Takeback
}

// NewMemoryStore returns a MemoryStore pointer
func NewMemoryStore() *MemoryStore {
	store := MemoryStore{
		games:      make(map[string]*Game, 10),
		challenges: make(map[string]*Challenge, 10),
	}
	return &store
}

// RetrieveGame will get a game from storage by its ID
func (m *MemoryStore) RetrieveGame(ID string) (*Game, error) {
	gm, ok := m.games[ID]
	if !ok {
		return nil, fmt.Errorf("Game by %v not found", ID)
	}
	return gm, nil
}

// StoreGame persists a game into memory
func (m *MemoryStore) StoreGame(ID string, game *Game) error {
	m.games[ID] = game
	return nil
}

// RetrieveChallenge will get a challenge request by challenger ID and challenged ID
func (m *MemoryStore) RetrieveChallenge(challengerID string, challengedID string) (*Challenge, error) {
	challenge, ok := m.challenges[challengerID+challengedID]
	if !ok {
		return nil, fmt.Errorf("Challenge %v%v not found", challengerID, challengedID)
	}
	return challenge, nil
}

// StoreChallenge will persist a challenge request
func (m *MemoryStore) StoreChallenge(c *Challenge) error {
	key := c.ChallengerID + c.ChallengedID
	m.challenges[key] = c
	return nil
}

// RemoveChallenge deletes a challenge request
func (m *MemoryStore) RemoveChallenge(challengerID string, challengedID string) error {
	key := challengerID + challengedID
	delete(m.challenges, key)
	return nil
}

// StoreTakeback stores a takeback request
func (m *MemoryStore) StoreTakeback(takeback *Takeback) error {
	key := takeback.CurrentGame.ID
	m.takebacks[key] = takeback
	return nil
}

// RetrieveTakeback finds a takeback request by a game ID
func (m *MemoryStore) RetrieveTakeback(gameID string) (*Takeback, error) {
	if takeback, ok := m.takebacks[gameID]; ok {
		return takeback, nil
	}
	return nil, fmt.Errorf("Takeback not found for provided game ID")
}

// RemoveTakeback removes a takeback request from storage
func (m *MemoryStore) RemoveTakeback(takeback *Takeback) error {
	delete(m.takebacks, takeback.CurrentGame.ID)
	return nil
}
