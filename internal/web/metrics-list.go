package web

import (
	"fmt"
	"html/template"
	"net/http"
)

var (
	templatePath = "../../internal/web/templates/list.tmpl"
)

func MetricsListPage(w http.ResponseWriter, metricsValues map[string]any) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("error parsing page template: %w", err)
	}

	err = tmpl.Execute(w, metricsValues)
	if err != nil {
		return fmt.Errorf("error filling page template: %w", err)
	}

	return nil
}
