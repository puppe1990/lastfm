package app

import (
	"github.com/puppe1990/amarra-cais/pkg/amarra/live"
	"github.com/puppe1990/amarra-cais/pkg/cais"

	"github.com/puppe1990/lastfm/internal/handlers"
)

func registerRoutes(r *cais.Router, deps Deps, cfg cais.Config) {
	listen := handlers.NewListeningHandler(deps.Views, deps.Library, deps.Site, deps.Catalog, cfg, deps.Store)
	r.Get("/", listen.Home)
	r.Get("/artistas", listen.Artists)
	r.Get("/albuns", listen.Albums)
	r.Get("/faixas", listen.Tracks)
	r.Get("/amadas", listen.Loved)
	r.Get("/robots.txt", listen.Robots)
}

func registerLiveViews(hub *live.Hub, deps Deps) {
	_ = hub
	_ = deps
	// cais:live-views
}
