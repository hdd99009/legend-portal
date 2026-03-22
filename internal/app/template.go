package app

import (
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	templates *template.Template
}

func NewTemplateRenderer(baseDir string) (*TemplateRenderer, error) {
	funcs := template.FuncMap{
		"safeHTML": func(content string) template.HTML {
			return template.HTML(content)
		},
		"hasID": func(ids []int64, target int64) bool {
			for _, id := range ids {
				if id == target {
					return true
				}
			}
			return false
		},
	}

	root := filepath.Join(baseDir, "templates")
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && filepath.Ext(path) == ".html" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("root").Funcs(funcs).ParseFiles(files...)
	if err != nil {
		return nil, err
	}

	return &TemplateRenderer{templates: tmpl}, nil
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}
