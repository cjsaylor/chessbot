package integration

type AuthStorage interface {
	StoreAuthToken(ID string, token string) error
	GetAuthToken(ID string) (string, error)
}
