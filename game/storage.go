package game

// GameStorage is an interface to be implemented for persisting a game
type GameStorage interface {
	RetrieveGame(ID string) (*Game, error)
	StoreGame(ID string, game *Game) error
}

// ChallengeStorage is an interface to be implemented for persisting challenges
type ChallengeStorage interface {
	RetrieveChallenge(challengerID string, challengedID string) (*Challenge, error)
	StoreChallenge(challenge *Challenge) error
	RemoveChallenge(challengerID string, challengedID string) error
}

// TakebackStorage is an interface to be implemented for persisting takeback requests.
type TakebackStorage interface {
	RetrieveTakeback(gameID string) (*Takeback, error)
	StoreTakeback(takeback *Takeback) error
	RemoveTakeback(takeback *Takeback) error
}
