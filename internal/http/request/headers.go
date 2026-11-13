package request // import "admin-console/internal/request"
import "net/http"

const (
	HTTP_HEADER_LOCALE string = "Locale"
)

func getHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

func GetLocale(r *http.Request) string {
	localeCode := getHeader(r, HTTP_HEADER_LOCALE)
	if localeCode == "" {
		localeCode = "zh_CN"
	}
	return localeCode
}
