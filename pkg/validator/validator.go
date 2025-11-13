package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	UserIDPattern   = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	MaxUserIDLength = 64
)

func ValidateUserID(userID string) error {
	if userID == "" {
		return nil
	}

	userID = strings.TrimSpace(userID)
	if len(userID) > MaxUserIDLength {
		return ErrUserIDTooLong
	}

	if !UserIDPattern.MatchString(userID) {
		return ErrInvalidUserIDFormat
	}

	return nil
}

func ValidateFilename(filename string) bool {
	if filename == "" {
		return false
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") || strings.Contains(filename, "\\") {
		return false
	}

	if utf8.RuneCountInString(filename) > 255 {
		return false
	}

	return true
}

func SanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "\\", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "..", "")
	if name == "" {
		return "file"
	}
	return name
}

func ValidateContentType(contentType string) bool {
	allowed := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
	}
	for _, ct := range allowed {
		if contentType == ct {
			return true
		}
	}
	return false
}
