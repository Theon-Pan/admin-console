package locale // import "admin-console/internal/locale"
import "errors"

type LocalizedError struct {
	translationKey  string
	translationArgs []any
}

func NewLocalizedError(translationKey string, translationArgs ...any) *LocalizedError {
	return &LocalizedError{translationKey: translationKey, translationArgs: translationArgs}
}

func (v *LocalizedError) String() string {
	return NewPrinter("en_US").Printf(v.translationKey, v.translationArgs...)
}

func (v *LocalizedError) Error() error {
	return errors.New(v.String())
}

func (v *LocalizedError) Translate(language string) string {
	return NewPrinter(language).Printf(v.translationKey, v.translationArgs...)
}
