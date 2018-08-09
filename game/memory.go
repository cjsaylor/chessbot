package game

import "fmt"

type MemoryStore struct {
	games map[string]*Game
}

func NewMemoryStore() *MemoryStore {
	store := MemoryStore{
		games: make(map[string]*Game, 10),
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
