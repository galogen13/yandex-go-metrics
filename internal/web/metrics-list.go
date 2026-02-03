package web

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
)

var (
	//go:embed templates/*.tmpl
	templateFS embed.FS
)

func MetricsListPage(metrics []*metrics.Metric) (bytes.Buffer, error) {

	var buf bytes.Buffer

	tmpl := template.Must(template.ParseFS(templateFS, "templates/list.tmpl"))

	err := tmpl.Execute(&buf, metrics)
	if err != nil {
		return buf, fmt.Errorf("error filling page template: %w", err)
	}

	return buf, nil
}
