package service

type AuthService interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
	GenerateAuthToken(userID string) (string, error)
	ValidateAuthToken(tokenStr string) (string, error)
}

