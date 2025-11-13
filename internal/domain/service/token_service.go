package service

type FileTokenClaims struct {
	FileID string  `json:"file_id"`
	Key    string  `json:"key"`
	UserID *string `json:"user_id,omitempty"`
}

type TokenService interface {
	Generate(fileID string, aesKey []byte, userID *string) (string, error)
	Validate(tokenStr string) (*FileTokenClaims, error)
}

