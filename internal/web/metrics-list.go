package web

import (
	"bytes"
	"fmt"
	"html/template"
)

var (
	templatePath = "internal/web/templates/list.tmpl"
)

func MetricsListPage(metricsValues map[string]any) (bytes.Buffer, error) {

	var buf bytes.Buffer

	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return buf, fmt.Errorf("error parsing page template: %w", err)
	}

	err = tmpl.Execute(&buf, metricsValues)
	if err != nil {
		return buf, fmt.Errorf("error filling page template: %w", err)
	}

	return buf, nil
}
