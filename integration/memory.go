package integration

import (
	"fmt"
)

// MemoryStore implements the Auth storage interfaces and holds all state in memory
// Once the MemoryStore instance is released, all data in that storage is lost
type MemoryStore struct {
	authorizations map[string]string
}

// NewMemoryStore returns a MemoryStore pointer
func NewMemoryStore() *MemoryStore {
	store := MemoryStore{
		authorizations: make(map[string]string, 10),
	}
	return &store
}

// StoreAuthToken stores the oauth token granted by slack user for a given team ID
func (m *MemoryStore) StoreAuthToken(teamID string, oauthToken string) error {
	m.authorizations[teamID] = oauthToken
	return nil
}

// GetAuthToken retrieves an oauth token for use with slack given a team ID
func (m *MemoryStore) GetAuthToken(teamID string) (string, error) {
	token, ok := m.authorizations[teamID]
	if !ok {
		return "", fmt.Errorf("Auth token not found for team %v", teamID)
	}
	return token, nil
}
