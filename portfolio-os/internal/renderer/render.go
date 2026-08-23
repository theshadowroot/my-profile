package renderer

import (
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
)

type Renderer struct {
	templates map[string]*template.Template
}

func New(files fs.FS) (*Renderer, error) {
	templates := make(map[string]*template.Template)

	funcs := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
	}

	basePages := map[string]string{
		"home":      "templates/pages/home.html",
		"client":    "templates/pages/client.html",
		"developer": "templates/pages/developer.html",
	}

	for name, page := range basePages {
		tmpl, err := template.New("").Funcs(funcs).ParseFS(
			files,
			"templates/layouts/base.html",
			page,
			"templates/partials/*.html",
		)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s (%s): %w", name, page, err)
		}

		templates[name] = tmpl
	}

	standalonePages := map[string]string{
		"admin":       "templates/pages/admin.html",
		"admin-login": "templates/pages/admin-login.html",
	}

	for name, page := range standalonePages {
		tmpl, err := template.New("").Funcs(funcs).ParseFS(files, page)
		if err != nil {
			return nil, fmt.Errorf("failed to parse standalone template %s: %w", name, err)
		}

		templates[name] = tmpl
	}

	return &Renderer{
		templates: templates,
	}, nil
}

func (r *Renderer) Render(
	w http.ResponseWriter,
	name string,
	data any,
) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template '%s' not found", name)
	}

	switch name {
	case "admin", "admin-login":
		return tmpl.ExecuteTemplate(w, name, data)
	default:
		return tmpl.ExecuteTemplate(w, "base", data)
	}
}