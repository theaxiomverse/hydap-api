// pkg/modules/cmd/cmd.go
package cmd

import (
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"hydap/pkg/modules/api"
	"hydap/pkg/modules/base"
	"hydap/pkg/modules/core"
	"log"
	"net/http"
)

func Run() {
	var defaultLoader base.ModuleLoader
	registry := core.NewModuleRegistry(defaultLoader)
	config, err := core.NewConfigManager("modules.db")
	if err != nil {
		log.Fatal(err)
	}

	metrics := core.NewMetricsExporter()
	api := api.NewModuleAPI(registry, config, metrics)

	r := chi.NewRouter()
	r.Mount("/api/v1", api.Router())
	r.Handle("/metrics", promhttp.Handler())

	// Hot reloader
	reloader, err := core.NewHotReloader(registry, log.Default())
	if err != nil {
		log.Fatal(err)
	}
	if err := reloader.WatchRecursive("./modules"); err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8080", r))
}
