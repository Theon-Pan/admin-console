package validator // import "admin-console/internal/validator"
import (
	"admin-console/internal/locale"
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"strings"
	"unicode"
)

// ValidateUserCreationWithPassword validates user creation with a password.
func ValidateUserCreationWithPassword(store *storage.Storage, request *model.UserCreationRequest) *locale.LocalizedError {
	if request.Username == "" {
		return locale.NewLocalizedError("error.user_mandatory_fields")
	}

	if store.UserExists(request.Username) {
		return locale.NewLocalizedError("error.user_already_exists")
	}

	if err := validateUsername(request.Username); err != nil {
		return err
	}

	if err := validatePassword(request.Password); err != nil {
		return err
	}

	return nil
}

func validatePassword(password string) *locale.LocalizedError {
	if len(password) < 6 {
		return locale.NewLocalizedError("error.password_min_length")
	}
	return nil
}

// validateUsername return an error if the `username` argument contains
// a character that isn't alphanumerical nor `_` and `-`.
//
// Note: this validation should not be applied to previously created usernames,
// and cannot be applied to Google/OIDC accounts creation because the email
// address is used for the username field.
func validateUsername(username string) *locale.LocalizedError {
	if strings.ContainsFunc(username, func(r rune) bool {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return false
		}
		if r == '_' || r == '-' || r == '@' || r == '.' {
			return false
		}
		return true
	}) {
		return locale.NewLocalizedError("error.invalid_username")
	}
	return nil
}
