package game

type GameStorage interface {
	RetrieveGame(ID string) (*Game, error)
	StoreGame(ID string, game *Game) error
}

type ChallengeStorage interface {
	RetrieveChallenge(challengerID string, challengedID string) (*Challenge, error)
	StoreChallenge(challenge *Challenge) error
	RemoveChallenge(challengerID string, challengedID string) error
}
