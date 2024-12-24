package core

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthEndpoint struct {
	registry *ModuleRegistry
}

func (h *HealthEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statuses := h.registry.GetAllHealth()
	json.NewEncoder(w).Encode(statuses)
}

func (r *ModuleRegistry) GetAllHealth() map[string]ModuleHealth {
	r.mu.RLock()
	defer r.mu.RUnlock()

	health := make(map[string]ModuleHealth)
	for name, mod := range r.modules {
		status := ModuleHealth{
			Status:      "healthy",
			LastChecked: time.Now(),
		}

		if err := mod.HealthCheck(); err != nil {
			status.Status = "unhealthy"
			status.Error = err.Error()
		}

		health[name] = status
	}
	return health
}

type ModuleHealth struct {
	Status      string    `json:"status"`
	LastChecked time.Time `json:"last_checked"`
	Error       string    `json:"error,omitempty"`
}
