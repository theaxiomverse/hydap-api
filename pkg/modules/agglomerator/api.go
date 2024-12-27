package agglomerator

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/theaxiomverse/hydap-api/pkg/modules/base"
	"net/http"
)

type API struct {
	module *AgglomeratorModule
}

func NewAPI(module *AgglomeratorModule) *API {
	return &API{module: module}
}

func (api *API) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/transaction", api.ProcessTransaction)
	r.Get("/chains", api.ListChains)
	r.Post("/chains", api.RegisterChain)
	r.Get("/chains/{id}", api.GetChain)
	r.Get("/status", api.GetStatus)
	r.Post("/pause", api.PauseModule)
	r.Post("/resume", api.ResumeModule)

	return r
}

// respondJSON is a helper function to send JSON responses
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// respondError is a helper function to send error responses
func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, map[string]string{"error": message})
}

func (api *API) ListChains(w http.ResponseWriter, r *http.Request) {
	agg := api.module.GetAgglomerator()
	if agg == nil {
		respondError(w, http.StatusServiceUnavailable, "agglomerator not initialized")
		return
	}

	chains := agg.ListChains()
	// Convert chains to a response format
	response := make([]map[string]interface{}, 0)
	for _, chain := range chains {
		chainData := map[string]interface{}{
			"id":       chain.ID,
			"endpoint": chain.Endpoint,
			"protocol": chain.Protocol,
		}
		response = append(response, chainData)
	}

	respondJSON(w, http.StatusOK, response)
}

func (api *API) GetChain(w http.ResponseWriter, r *http.Request) {
	chainID := chi.URLParam(r, "id")
	agg := api.module.GetAgglomerator()
	if agg == nil {
		respondError(w, http.StatusServiceUnavailable, "agglomerator not initialized")
		return
	}

	chain, err := agg.GetChain(chainID)
	if err != nil {
		respondError(w, http.StatusNotFound, "chain not found")
		return
	}

	response := map[string]interface{}{
		"id":       chain.ID,
		"endpoint": chain.Endpoint,
		"protocol": chain.Protocol,
	}

	respondJSON(w, http.StatusOK, response)
}

func (api *API) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"state":   api.module.GetState().String(),
		"health":  api.module.HealthCheck() == nil,
		"version": api.module.Version(),
		"config":  api.module.GetConfig(),
	}

	respondJSON(w, http.StatusOK, status)
}

func (api *API) ProcessTransaction(w http.ResponseWriter, r *http.Request) {
	var tx Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := api.module.ProcessTransaction(&tx); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"id":     tx.ID,
		"status": "accepted",
	}
	respondJSON(w, http.StatusAccepted, response)
}

func (api *API) RegisterChain(w http.ResponseWriter, r *http.Request) {
	var chain Chain
	if err := json.NewDecoder(r.Body).Decode(&chain); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	agg := api.module.GetAgglomerator()
	if agg == nil {
		respondError(w, http.StatusServiceUnavailable, "agglomerator not initialized")
		return
	}

	if err := agg.RegisterChain(&chain); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := map[string]interface{}{
		"id":      chain.ID,
		"status":  "registered",
		"message": "chain successfully registered",
	}
	respondJSON(w, http.StatusCreated, response)
}

func (api *API) PauseModule(w http.ResponseWriter, r *http.Request) {
	if api.module.GetState() != base.StateRunning {
		respondError(w, http.StatusBadRequest, "module not running")
		return
	}

	api.module.SetState(base.StatePaused)
	respondJSON(w, http.StatusOK, map[string]string{"status": "paused"})
}

func (api *API) ResumeModule(w http.ResponseWriter, r *http.Request) {
	if api.module.GetState() != base.StatePaused {
		respondError(w, http.StatusBadRequest, "module not paused")
		return
	}

	api.module.SetState(base.StateRunning)
	respondJSON(w, http.StatusOK, map[string]string{"status": "running"})
}
