package agglomerator

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/theaxiomverse/hydap-api/pkg/modules/base"
	"net/http"
)

type AgglomeratorModule struct {
	base.Module
}

type API struct {
	module *AgglomeratorModule
}

func NewAPI(module *AgglomeratorModule) *API {
	return &API{module: module}
}

func (api *API) Routes() chi.Router {
	r := chi.NewRouter()

	// Transaction endpoints
	r.Post("/transaction", api.ProcessTransaction)

	// Chain management
	r.Get("/chains", api.ListChains)
	r.Post("/chains", api.RegisterChain)
	r.Get("/chains/{id}", api.GetChain)

	// Module management
	r.Get("/status", api.GetStatus)
	r.Post("/pause", api.PauseModule)
	r.Post("/resume", api.ResumeModule)

	return r
}

func (api *API) ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	var tx Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := api.module.ProcessTransaction(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (api *API) ListChains(w http.ResponseWriter, r *http.Request) {
	agglomerator := api.module.GetAgglomerator()
	if agglomerator == nil {
		http.Error(w, "agglomerator not initialized", http.StatusServiceUnavailable)
		return
	}

	chains := agglomerator.ListChains()
	json.NewEncoder(w).Encode(chains)
}

func (api *API) GetChain(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "id")
	agglomerator := api.module.GetAgglomerator()
	if agglomerator == nil {
		http.Error(w, "agglomerator not initialized", http.StatusServiceUnavailable)
		return
	}

	chain, err := agglomerator.GetChain(chainID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(chain)
}

func (api *API) RegisterChain(w http.ResponseWriter, r *http.Request) {
	var chain Chain
	if err := json.NewDecoder(r.Body).Decode(&chain); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	agglomerator := api.module.GetAgglomerator()
	if agglomerator == nil {
		http.Error(w, "agglomerator not initialized", http.StatusServiceUnavailable)
		return
	}

	if err := agglomerator.RegisterChain(&chain); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (api *API) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"state":   api.module.State().String(),
		"health":  api.module.HealthCheck() == nil,
		"version": api.module.Version(),
		"config":  api.module.config,
	}
	json.NewEncoder(w).Encode(status)
}

func (api *API) PauseModule(w http.ResponseWriter, r *http.Request) {
	if api.module.State() != base.StateRunning {
		http.Error(w, "module not running", http.StatusBadRequest)
		return
	}

	api.module.state = base.StatePaused
	w.WriteHeader(http.StatusOK)
}

func (api *API) ResumeModule(w http.ResponseWriter, r *http.Request) {
	if api.module.State() != base.StatePaused {
		http.Error(w, "module not paused", http.StatusBadRequest)
		return
	}

	api.module.state = base.StateRunning
	w.WriteHeader(http.StatusOK)
}
