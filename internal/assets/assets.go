package assets

import (
	"embed"
	"html/template"
	"io/fs"
)

//go:embed web/templates/*
var TemplateFS embed.FS

//go:embed web/static/*
var StaticFS embed.FS

// GetTemplateFS returns the embedded template filesystem
func GetTemplateFS() fs.FS {
	return TemplateFS
}

// GetStaticFS returns the embedded static filesystem
func GetStaticFS() fs.FS {
	return StaticFS
}

// ParseTemplates parses all embedded templates with the given function map
func ParseTemplates(funcMap template.FuncMap) (*template.Template, error) {
	return template.New("").Funcs(funcMap).ParseFS(TemplateFS, "web/templates/*.html")
}