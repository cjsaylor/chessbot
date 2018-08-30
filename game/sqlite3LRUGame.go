package game

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cjsaylor/goutil/lru"
)

type SqliteLRUGameStore struct {
	base  *SqliteStore
	cache *lru.Cache
	trap  chan os.Signal
}

func NewSqliteLRUStore(base *SqliteStore, capacity uint, onFinishedCleanup func()) (*SqliteLRUGameStore, error) {
	store := &SqliteLRUGameStore{base: base}
	evictionHandler := func(key, value interface{}) {
		store.base.StoreGame(key.(string), value.(*Game))
	}
	store.cache = lru.NewCache(capacity, evictionHandler)
	// Ensure that we store the games when the receiving a kill signal
	store.trap = make(chan os.Signal, 2)
	signal.Notify(store.trap, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-store.trap
		for _, ID := range store.cache.ListKeys() {
			game, _ := store.cache.Get(ID)
			store.base.StoreGame(ID.(string), game.(*Game))
		}
		onFinishedCleanup()
	}()
	return store, nil
}

func (s *SqliteLRUGameStore) RetrieveGame(ID string) (*Game, error) {
	if game, ok := s.cache.Get(ID); ok {
		return game.(*Game), nil
	}
	return s.base.RetrieveGame(ID)
}

func (s *SqliteLRUGameStore) StoreGame(ID string, game *Game) error {
	s.cache.Set(ID, game)
	return nil
}
