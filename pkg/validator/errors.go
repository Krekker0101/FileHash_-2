package validator

import "errors"

var (
	ErrUserIDTooLong        = errors.New("user_id exceeds maximum length")
	ErrInvalidUserIDFormat  = errors.New("user_id contains invalid characters")
	ErrInvalidFilename      = errors.New("invalid filename")
	ErrInvalidContentType    = errors.New("unsupported content type")
)

