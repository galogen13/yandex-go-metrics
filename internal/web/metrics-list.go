package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

var (
	//templatePath = "internal/web/templates/list.tmpl"

	//go:embed templates/*.tmpl
	templateFS embed.FS
)

func MetricsListPage(metricsValues map[string]any) (bytes.Buffer, error) {

	var buf bytes.Buffer

	tmpl := template.Must(template.ParseFS(templateFS, "templates/list.tmpl"))

	err := tmpl.Execute(&buf, metricsValues)
	if err != nil {
		return buf, fmt.Errorf("error filling page template: %w", err)
	}

	return buf, nil
}
