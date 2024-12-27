// pkg/modules/api/handlers.go
package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
"github.com/theaxiomverse/hydap-api/base"
"github.com/theaxiomverse/hydap-api/pkg/modules/core"
"net/http"
)
type ModuleAPI struct {
	registry *core.ModuleRegistry
	config   *core.ConfigManager
	metrics  *core.MetricsExporter
}

func NewModuleAPI(registry *core.ModuleRegistry, config *core.ConfigManager, metrics *core.MetricsExporter) *ModuleAPI {
	return &ModuleAPI{
		registry: registry,
		config:   config,
		metrics:  metrics,
	}
}

func (api *ModuleAPI) ListModules(w http.ResponseWriter, r *http.Request) {
	modules := api.registry.List()
	json.NewEncoder(w).Encode(modules)
}

func (api *ModuleAPI) GetHealth(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	health := api.registry.GetAllHealth()[name]
	json.NewEncoder(w).Encode(health)
}

func (api *ModuleAPI) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var config json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := api.config.SetConfig(name, config); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (api *ModuleAPI) AddModule(w http.ResponseWriter, r *http.Request) {
	var config base.ModuleConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mod, err := api.registry.Loader.LoadFromConfig(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := api.registry.RegisterWithDeps(mod, config.DependsOn); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	api.metrics.RegisterModule(mod.Name())
	w.WriteHeader(http.StatusCreated)
}

func (api *ModuleAPI) GetModule(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	mod, exists := api.registry.Get(name)
	if !exists {
		http.Error(w, "module not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(mod)
}

func (api *ModuleAPI) DeleteModule(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := api.registry.Terminate(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *ModuleAPI) StartModule(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	mod, exists := api.registry.Get(name)
	if !exists {
		http.Error(w, "module not found", http.StatusNotFound)
		return
	}

	if err := mod.Initialize(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (api *ModuleAPI) StopModule(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	mod, exists := api.registry.Get(name)
	if !exists {
		http.Error(w, "module not found", http.StatusNotFound)
		return
	}

	if err := mod.Terminate(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
