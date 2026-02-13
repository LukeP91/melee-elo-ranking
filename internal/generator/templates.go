package generator

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"path/filepath"
	"sync"
)

//go:embed templates/*.tmpl templates/partials/*.tmpl
var embedFS embed.FS

var (
	templates   *template.Template
	templatesOnce sync.Once
)

func getTemplates() *template.Template {
	templatesOnce.Do(func() {
		templates = template.Must(template.New("").ParseFS(embedFS, "templates/*.tmpl", "templates/partials/*.tmpl"))
	})
	return templates
}

// executeTemplate executes the named template. Tries both "templates/name" and "name"
// since ParseFS template names can vary by Go version and embed path.
func executeTemplate(name string, data interface{}) ([]byte, error) {
	tmpl := getTemplates()
	t := tmpl.Lookup(name)
	if t == nil {
		// Try base name (e.g. "index.tmpl") in case ParseFS registered by filename only
		t = tmpl.Lookup(filepath.Base(name))
	}
	if t == nil {
		return nil, fmt.Errorf("html/template: %q is undefined", name)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
