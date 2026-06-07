package views

import (
	"html/template"
	"net/http"
	"path/filepath"
)

type Renderer struct {
	base string
}

func New(base string) *Renderer { return &Renderer{base: base} }

func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
	files := []string{
		filepath.Join(r.base, "layout.html"),
		filepath.Join(r.base, "partials", "flash.html"),
		filepath.Join(r.base, name),
	}
	t, err := template.ParseFiles(files...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
