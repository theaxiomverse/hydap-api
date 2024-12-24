package api

import "github.com/go-chi/chi/v5"

func (api *ModuleAPI) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/modules", api.ListModules)
	r.Post("/modules", api.AddModule)
	r.Route("/modules/{name}", func(r chi.Router) {
		r.Get("/", api.GetModule)
		r.Get("/health", api.GetHealth)
		r.Put("/config", api.UpdateConfig)
		r.Delete("/", api.DeleteModule)
		r.Post("/start", api.StartModule)
		r.Post("/stop", api.StopModule)
	})

	return r
}
