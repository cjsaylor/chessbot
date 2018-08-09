package game

type GameStorage interface {
	RetrieveGame(id string) (*Game, error)
	StoreGame(id string, game *Game) error
}
