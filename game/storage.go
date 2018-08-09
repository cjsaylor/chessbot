package game

type GameStorage interface {
	RetrieveGame(id string) (*Game, error)
	StoreGame(id string, game *Game) error
}

type ChallengeStorage interface {
	RetrieveChallenge(challengerID string, challengedID string) (*Challenge, error)
	StoreChallenge(challenge *Challenge) error
}
