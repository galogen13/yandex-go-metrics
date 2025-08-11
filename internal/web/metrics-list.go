package web

import (
	"html/template"
	"net/http"
)

var (
	templatePath = "../../internal/web/templates/list.tmpl"
)

func MetricsListPage(w http.ResponseWriter, metricsValues map[string]any) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(w, metricsValues)
	if err != nil {
		return err
	}

	return nil
}
