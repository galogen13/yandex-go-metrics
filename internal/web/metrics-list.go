package web

import (
	"html/template"
	"net/http"
)

var (
	templatePath = "../../internal/web/templates/list.tmpl"
)

func MetricsListPage(w http.ResponseWriter, metricsValues map[string]any) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, metricsValues)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
