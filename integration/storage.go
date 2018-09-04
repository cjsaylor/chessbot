package integration

// AuthStorage interface guarentees implemented mmethods for oauth token storage
type AuthStorage interface {
	StoreAuthToken(ID string, token string) error
	GetAuthToken(ID string) (string, error)
}
