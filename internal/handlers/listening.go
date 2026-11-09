package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/meta"

	"github.com/puppe1990/lastfm/internal/lastfm"
)

type ListeningHandler struct {
	views   *view.Renderer
	lib     lastfm.Library
	site    meta.Site
	catalog *i18n.Catalog
	cfg     cais.Config
	now     func() time.Time
}

func NewListeningHandler(views *view.Renderer, lib lastfm.Library, site meta.Site, catalog *i18n.Catalog, cfg cais.Config) *ListeningHandler {
	return &ListeningHandler{views: views, lib: lib, site: site, catalog: catalog, cfg: cfg, now: time.Now}
}

func (h *ListeningHandler) Home(w http.ResponseWriter, r *http.Request) {
	pageN := queryPage(r)
	user, err := h.lib.User(r.Context())
	if err != nil {
		h.fail(w, r, "home", err)
		return
	}
	recent, err := h.lib.RecentTracks(r.Context(), pageN, 20)
	if err != nil {
		h.fail(w, r, "home", err)
		return
	}
	window, werr := h.lib.RecentTracks(r.Context(), 1, 200)
	if werr != nil {
		window = recent
	}
	week, _ := h.lib.TopArtists(r.Context(), "7day", 1, 8)
	h.write(w, r, "home", h.t(r, "home.title"), "home", http.StatusOK, presentHome(user, recent, window, week, pageN, h.clock()))
}

func (h *ListeningHandler) Artists(w http.ResponseWriter, r *http.Request) {
	period := lastfm.NormalizePeriod(r.URL.Query().Get("period"))
	pageN := queryPage(r)
	chart, err := h.lib.TopArtists(r.Context(), period, pageN, 20)
	if err != nil {
		h.fail(w, r, "artists", err)
		return
	}
	h.write(w, r, "artists", h.t(r, "artists.title"), "artists", http.StatusOK, presentArtists(chart, period, pageN))
}

func (h *ListeningHandler) Albums(w http.ResponseWriter, r *http.Request) {
	period := lastfm.NormalizePeriod(r.URL.Query().Get("period"))
	pageN := queryPage(r)
	chart, err := h.lib.TopAlbums(r.Context(), period, pageN, 20)
	if err != nil {
		h.fail(w, r, "albums", err)
		return
	}
	h.write(w, r, "albums", h.t(r, "albums.title"), "albums", http.StatusOK, presentAlbums(chart, period, pageN))
}

func (h *ListeningHandler) Tracks(w http.ResponseWriter, r *http.Request) {
	period := lastfm.NormalizePeriod(r.URL.Query().Get("period"))
	pageN := queryPage(r)
	chart, err := h.lib.TopTracks(r.Context(), period, pageN, 20)
	if err != nil {
		h.fail(w, r, "tracks", err)
		return
	}
	h.write(w, r, "tracks", h.t(r, "tracks.title"), "tracks", http.StatusOK, presentTracks(chart, period, pageN))
}

func (h *ListeningHandler) Loved(w http.ResponseWriter, r *http.Request) {
	pageN := queryPage(r)
	chart, err := h.lib.LovedTracks(r.Context(), pageN, 20)
	if err != nil {
		h.fail(w, r, "loved", err)
		return
	}
	h.write(w, r, "loved", h.t(r, "loved.title"), "loved", http.StatusOK, presentLoved(chart, pageN))
}

func (h *ListeningHandler) Robots(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "User-agent: *\nDisallow: /\n")
}

func (h *ListeningHandler) fail(w http.ResponseWriter, r *http.Request, name string, err error) {
	log.Printf("last.fm %s: %v", name, err)
	h.write(w, r, name, h.t(r, name+".title"), name, http.StatusBadGateway, listenPage{
		Error: h.t(r, "listen.error"),
	})
}

func (h *ListeningHandler) write(w http.ResponseWriter, r *http.Request, name, title, nav string, status int, page listenPage) {
	data := amarraData(r, h.site, map[string]any{
		"Title":      title,
		"ActiveNav":  nav,
		"ShowPeriod": page.ShowPeriod,
		"Period":     page.Period,
		"Periods":    page.Periods,
		"Page":       page,
	})
	writeView(w, r, h.views, h.cfg, "app", name, data, status)
}

func (h *ListeningHandler) t(r *http.Request, key string, args ...any) string {
	return i18n.CatalogOr(r, h.catalog).T(key, args...)
}

func (h *ListeningHandler) clock() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

func queryPage(r *http.Request) int {
	n, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if n < 1 {
		return 1
	}
	return n
}
