package main

import (
	"Test/handler"
	"github.com/labstack/echo"
	"html/template"
	"io"
)
// Define the template registry struct
type TemplateRegistry struct {
	templates *template.Template
}

// Implement e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	// Instantiate a template registry and register all html files inside the view folder

	e.Renderer = &TemplateRegistry{
		templates: template.Must(template.ParseGlob("static/*.html")),
	}
	//e.Use(middleware.BasicAuth(Log))
	e.GET("/insert", handler.Insert)
	e.GET("/login", handler.Login)
	e.GET("/show", handler.Show)
	e.GET("/del", handler.Delete)
	e.Start(":8000")
}
