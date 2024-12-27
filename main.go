package main

import (
	"github.com/go-chi/chi/v5"

	"github.com/theaxiomverse/hydap-api/pkg/modules/agglomerator"
	"github.com/theaxiomverse/hydap-api/pkg/modules/api"
	"github.com/theaxiomverse/hydap-api/pkg/modules/base"
	"github.com/theaxiomverse/hydap-api/pkg/modules/core"
	"log"
	"net/http"
	"os"
)

func main() {
	// Initialize core components
	configManager, err := core.NewConfigManager("./config.db")
	if err != nil {
		log.Fatal(err)
	}

	metrics := core.NewMetricsExporter()
	logger := &core.ModuleLogger{
		Outputs: make(map[string]*os.File),
	}

	// Or create it via the loader system
	loader := agglomerator.NewAgglomeratorLoader(configManager, metrics, logger)
	registry := core.NewModuleRegistry(loader)

	// Create module config
	moduleConfig := base.ModuleConfig{
		Name:      "blockchain_agglomerator",
		Version:   "1.0.0",
		DependsOn: []string{},
		Config: map[string]interface{}{
			"nodeID":        "node1",
			"vectorDims":    50,
			"simThreshold":  0.7,
			"enabledChains": []string{"ethereum", "solana"},
			"logPath":       "./agglomerator.log",
		},
	}

	// Load and register module
	module, err := loader.LoadFromConfig(moduleConfig)
	if err != nil {
		log.Fatal(err)
	}

	if err := registry.RegisterWithDeps(module, moduleConfig.DependsOn); err != nil {
		log.Fatal(err)
	}

	// Create API handlers
	moduleAPI := api.NewModuleAPI(registry, configManager, metrics)

	// Setup router
	r := chi.NewRouter()
	r.Mount("/api/modules", moduleAPI.Router())
	//	r.Mount("/api/agglomerator", api.NewModuleAPI((module.(*agglomerator.AgglomeratorModule)).Routes())

	// Setup hot reloading
	hotReloader, err := core.NewHotReloader(registry, log.New(os.Stdout, "", log.LstdFlags))
	if err != nil {
		log.Fatal(err)
	}

	if err := hotReloader.WatchRecursive("./modules"); err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":8080", r))
}
