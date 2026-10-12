package template // import "admin-console/internal/template"
import (
	"admin-console/internal/model"
	"html/template"
	"strings"
)

type funcMap struct {
	basePath  string
	iconPaths map[string]string
}

// Map returns a map of template functions that are compiled during template parsing.
func (f *funcMap) Map() template.FuncMap {
	return template.FuncMap{
		"csp": csp,
	}
}

func csp(user *model.User, nonce string) string {
	policies := map[string]string{
		"default-src":               "'none'",
		"frame-src":                 "*",
		"img-src":                   "* data:",
		"manifest-src":              "'self'",
		"media-src":                 "*",
		"require-trusted-types-for": "'script'",
		"script-src":                "'nonce-" + nonce + "' 'strict-dynamic'",
		"style-src":                 "'nonce-" + nonce + "'",
		"trusted-types":             "html url",
		"connect-src":               "'self'",
	}

	var policy strings.Builder
	policy.Grow(350)
	for key, value := range policies {
		policy.WriteString(key)
		policy.WriteString(" ")
		policy.WriteString(value)
		policy.WriteString("; ")
	}

	return `<meta http-equiv="Content-Security-Policy content="` + policy.String() + `">`

}
