package renderer

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/ViBiOh/httputils/v3/pkg/flags"
	"github.com/ViBiOh/httputils/v3/pkg/httperror"
	"github.com/ViBiOh/httputils/v3/pkg/templates"
	"github.com/ViBiOh/hue/pkg/renderer/model"
)

const (
	faviconPath = "/favicon"
	svgPath     = "/svg"
)

var (
	rootPaths = []string{"/robots.txt", "/sitemap.xml"}
	staticDir = "static"
)

// App of package
type App interface {
	Handler(model.TemplateFunc) http.Handler
	Error(http.ResponseWriter, error)
	Redirect(http.ResponseWriter, *http.Request, string, string)
}

// Config of package
type Config struct {
	templates *string
	statics   *string
	publicURL *string
	title     *string
}

type app struct {
	tpl        *template.Template
	staticsDir string
	content    map[string]interface{}
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string, overrides ...flags.Override) Config {
	return Config{
		templates: flags.New(prefix, "").Name("Templates").Default(flags.Default("Templates", "./templates/", overrides)).Label("HTML Templates folder").ToString(fs),
		statics:   flags.New(prefix, "").Name("Static").Default(flags.Default("Static", "./static/", overrides)).Label("Static folder, content served directly").ToString(fs),
		publicURL: flags.New(prefix, "").Name("PublicURL").Default(flags.Default("PublicURL", "http://localhost", overrides)).Label("Public URL").ToString(fs),
		title:     flags.New(prefix, "").Name("Title").Default(flags.Default("Title", "App", overrides)).Label("Application title").ToString(fs),
	}
}

// New creates new App from Config
func New(config Config, funcMap template.FuncMap) (App, error) {
	filesTemplates, err := templates.GetTemplates(strings.TrimSpace(*config.templates), ".html")
	if err != nil {
		return nil, fmt.Errorf("unable to get templates: %s", err)
	}

	return app{
		tpl: template.Must(template.New("app").Funcs(funcMap).ParseFiles(filesTemplates...)),
		content: map[string]interface{}{
			"PublicURL": strings.TrimSpace(*config.publicURL),
			"Title":     strings.TrimSpace(*config.title),
			"Version":   os.Getenv("VERSION"),
		},
	}, nil
}

func isRootPaths(requestPath string) bool {
	for _, rootPath := range rootPaths {
		if strings.EqualFold(rootPath, requestPath) {
			return true
		}
	}

	return false
}

func (a app) feedContent(content map[string]interface{}) {
	for key, value := range a.content {
		content[key] = value
	}
}

func (a app) Handler(templateFunc model.TemplateFunc) http.Handler {
	svgHandler := http.StripPrefix(svgPath, a.svg())

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, faviconPath) || isRootPaths(r.URL.Path) {
			http.ServeFile(w, r, path.Join(a.staticsDir, r.URL.Path))
			return
		}

		if a.tpl == nil {
			httperror.NotFound(w)
			return
		}

		if strings.HasPrefix(r.URL.Path, svgPath) {
			svgHandler.ServeHTTP(w, r)
			return
		}

		templateName, status, content, err := templateFunc(r)
		if err != nil {
			a.Error(w, err)
			return
		}

		a.feedContent(content)

		message := model.ParseMessage(r)
		if len(message.Content) > 0 {
			content["Message"] = message
		}

		if err := templates.ResponseHTMLTemplate(a.tpl.Lookup(templateName), w, content, status); err != nil {
			httperror.InternalServerError(w, err)
		}
	})
}
