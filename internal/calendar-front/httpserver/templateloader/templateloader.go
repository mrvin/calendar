package templateloader

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"time"
)

//go:embed templates/*
var files embed.FS

type TemplateLoader struct {
	templates map[string]*template.Template
}

func New() *TemplateLoader {
	return &TemplateLoader{make(map[string]*template.Template)}
}

func dateFormat(layout string, d time.Time) string {
	return d.Format(layout)
}

func (t *TemplateLoader) Load(templatesDir string) {
	t.templates["form-user.html"] = template.Must(template.ParseFS(files,
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "form-user.html")))
	t.templates["form-login.html"] = template.Must(template.ParseFS(files,
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "form-login.html")))
	t.templates["form-event.html"] = template.Must(template.ParseFS(files,
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "form-event.html")))
	t.templates["text.html"] = template.Must(template.ParseFS(files,
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "text.html")))
	t.templates["list-users.html"] = template.Must(template.ParseFS(files,
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "list-users.html")))

	name := "base.html"
	t.templates["list-events.html"] = template.Must(
		template.New(name).
			Funcs(template.FuncMap{"dateFormat": dateFormat}).
			ParseFS(files,
				filepath.Join(templatesDir, name),
				filepath.Join(templatesDir, "list-events.html"),
			),
	)
	t.templates["user.html"] = template.Must(template.ParseFS(files,
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "user.html")))

	name = "base.html"
	t.templates["event.html"] = template.Must(
		template.New(name).
			Funcs(template.FuncMap{"dateFormat": dateFormat}).
			ParseFS(
				files,
				filepath.Join(templatesDir, name),
				filepath.Join(templatesDir, "event.html"),
			),
	)
}

func (t *TemplateLoader) Execute(nameTemp string, res http.ResponseWriter, data any) error {
	temp, ok := t.templates[nameTemp]
	if !ok {
		return fmt.Errorf("not found template '%s'", nameTemp)
	}

	if err := temp.Execute(res, data); err != nil {
		return fmt.Errorf("execute %w", err)
	}

	return nil
}
