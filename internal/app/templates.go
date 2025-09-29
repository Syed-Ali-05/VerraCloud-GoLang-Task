package app

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

//go:embed templates/*.templ
var tmplFS embed.FS

// mustParseSet parses the given template files with shared helper functions.
func mustParseSet(files ...string) *template.Template {
	t := template.New("").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"formatTime": func(ts any) string {
			switch v := ts.(type) {
			case time.Time:
				return v.Format("2006-01-02 15:04")
			case int64:
				return time.Unix(v, 0).Format("2006-01-02 15:04")
			default:
				return ""
			}
		},
	})
	tt, err := t.ParseFS(tmplFS, files...)
	if err != nil {
		log.Fatalf("template parse: %v", err)
	}
	return tt
}

var (
	tmplBaseLogin     = mustParseSet("templates/base.templ", "templates/login.templ")
	tmplBaseDashboard = mustParseSet("templates/base.templ", "templates/dashboard.templ", "templates/items.templ")
)

var (
	tmplLoginPartial     = mustParseSet("templates/login.templ")
	tmplDashboardPartial = mustParseSet("templates/dashboard.templ")
	tmplItemsPartial     = mustParseSet("templates/items.templ")
)

// mustParseSet parses the given template files with shared helper functions.
func renderLoginPartial(w http.ResponseWriter, email, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmplLoginPartial.ExecuteTemplate(w, "login.templ", map[string]any{
		"Email": email,
		"Error": errMsg,
	}); err != nil {
		httpErrorFragment(w, err)
	}
}
