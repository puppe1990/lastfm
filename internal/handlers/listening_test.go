package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/fakedata"

	appi18n "github.com/puppe1990/lastfm/internal/i18n"
	"github.com/puppe1990/lastfm/internal/lastfm"
)

type stubLibrary struct {
	user    lastfm.User
	recent  lastfm.RecentPage
	artists lastfm.ArtistPage
	albums  lastfm.AlbumPage
	tracks  lastfm.TrackPage
	loved   lastfm.TrackPage
	err     error
	period  string
}

func (s *stubLibrary) UserName() string { return s.user.Name }

func (s *stubLibrary) User(context.Context) (lastfm.User, error) {
	if s.err != nil {
		return lastfm.User{}, s.err
	}
	return s.user, nil
}

func (s *stubLibrary) RecentTracks(context.Context, int, int) (lastfm.RecentPage, error) {
	if s.err != nil {
		return lastfm.RecentPage{}, s.err
	}
	return s.recent, nil
}

func (s *stubLibrary) TopArtists(_ context.Context, period string, _, _ int) (lastfm.ArtistPage, error) {
	s.period = period
	if s.err != nil {
		return lastfm.ArtistPage{}, s.err
	}
	return s.artists, nil
}

func (s *stubLibrary) TopAlbums(_ context.Context, period string, _, _ int) (lastfm.AlbumPage, error) {
	s.period = period
	if s.err != nil {
		return lastfm.AlbumPage{}, s.err
	}
	return s.albums, nil
}

func (s *stubLibrary) TopTracks(_ context.Context, period string, _, _ int) (lastfm.TrackPage, error) {
	s.period = period
	if s.err != nil {
		return lastfm.TrackPage{}, s.err
	}
	return s.tracks, nil
}

func (s *stubLibrary) LovedTracks(context.Context, int, int) (lastfm.TrackPage, error) {
	if s.err != nil {
		return lastfm.TrackPage{}, s.err
	}
	return s.loved, nil
}

func seedLibrary(t *testing.T) *stubLibrary {
	t.Helper()
	fakedata.Seed(42)
	now := lastfm.Track{
		Name:       fakedata.Title(),
		Artist:     fakedata.Name(),
		Album:      fakedata.Title(),
		URL:        "https://www.last.fm/music/now",
		Image:      "https://img.example/now.jpg",
		NowPlaying: true,
	}
	past := lastfm.Track{
		Name:       fakedata.Title(),
		Artist:     fakedata.Name(),
		Album:      fakedata.Title(),
		URL:        "https://www.last.fm/music/past",
		Image:      "https://img.example/past.jpg",
		PlayedUnix: 1731167280,
	}
	artist := lastfm.Artist{Name: fakedata.Name(), Playcount: 12, Rank: 1, URL: "https://www.last.fm/music/a"}
	album := lastfm.Album{Name: fakedata.Title(), Artist: fakedata.Name(), Playcount: 9, Rank: 1, Image: "https://img.example/al.jpg"}
	loved := lastfm.Track{Name: fakedata.Title(), Artist: fakedata.Name(), PlayedUnix: 1600000000, URL: "https://www.last.fm/music/loved"}
	return &stubLibrary{
		user: lastfm.User{
			Name:           "puppe90",
			RealName:       "Matheus Nunes Puppe",
			URL:            "https://www.last.fm/user/puppe90",
			Playcount:      82197,
			ArtistCount:    2957,
			AlbumCount:     9017,
			TrackCount:     23622,
			RegisteredUnix: 1201408060,
			Image:          "https://img.example/me.png",
		},
		recent:  lastfm.RecentPage{NowPlaying: &now, Tracks: []lastfm.Track{past}, Page: 1, TotalPages: 2, Total: 82197},
		artists: lastfm.ArtistPage{Artists: []lastfm.Artist{artist}, Page: 1, TotalPages: 1, Total: 1},
		albums:  lastfm.AlbumPage{Albums: []lastfm.Album{album}, Page: 1, TotalPages: 1, Total: 1},
		tracks:  lastfm.TrackPage{Tracks: []lastfm.Track{past}, Page: 1, TotalPages: 1, Total: 1},
		loved:   lastfm.TrackPage{Tracks: []lastfm.Track{loved}, Page: 1, TotalPages: 1, Total: 1},
	}
}

func newListening(t *testing.T, lib lastfm.Library) *ListeningHandler {
	t.Helper()
	return NewListeningHandler(setupTestViews(t), lib, testSite(), appi18n.NewCatalog("pt"), cais.Config{})
}

func TestHome_ShowsProfileNowPlayingAndRecent(t *testing.T) {
	lib := seedLibrary(t)
	h := newListening(t, lib)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.Home(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `id="amarra-main"`) {
		t.Fatal("missing #amarra-main")
	}
	if !strings.Contains(body, "puppe90") {
		t.Fatal("missing username")
	}
	if !strings.Contains(body, "82.197") {
		t.Fatal("missing grouped playcount")
	}
	if !strings.Contains(body, lib.recent.NowPlaying.Name) {
		t.Fatalf("missing now playing %q", lib.recent.NowPlaying.Name)
	}
	if !strings.Contains(body, lib.recent.Tracks[0].Name) {
		t.Fatalf("missing recent %q", lib.recent.Tracks[0].Name)
	}
	if strings.Contains(body, "is running") {
		t.Fatal("scaffold heading still present")
	}
	if !strings.Contains(body, `href="/artistas"`) {
		t.Fatal("missing artistas nav")
	}
}

func TestHome_ShowsLastfmStyleCharts(t *testing.T) {
	lib := seedLibrary(t)
	loc := time.FixedZone("BRT", -3*3600)
	now := time.Date(2026, 11, 9, 18, 0, 0, 0, loc)
	lib.recent.Tracks = []lastfm.Track{
		{Name: "Bleed", Artist: "Meshuggah", PlayedUnix: time.Date(2026, 11, 9, 17, 0, 0, 0, loc).Unix()},
		{Name: "obZen", Artist: "Meshuggah", PlayedUnix: time.Date(2026, 11, 8, 12, 0, 0, 0, loc).Unix()},
	}
	lib.artists.Artists = []lastfm.Artist{
		{Name: "Trivium", Playcount: 12, Rank: 1, URL: "https://www.last.fm/music/Trivium"},
		{Name: "Meshuggah", Playcount: 6, Rank: 2},
	}
	h := newListening(t, lib)
	h.now = func() time.Time { return now }

	rr := httptest.NewRecorder()
	h.Home(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{"Atividade", "Relógio", "listen-activity", "listen-clock", "Trivium", "listen-share", ">seg<"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestArtists_ListsTopAndKeepsPeriod(t *testing.T) {
	lib := seedLibrary(t)
	h := newListening(t, lib)

	req := httptest.NewRequest(http.MethodGet, "/artistas?period=1month", nil)
	rr := httptest.NewRecorder()
	h.Artists(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if lib.period != "1month" {
		t.Errorf("period = %q", lib.period)
	}
	body := rr.Body.String()
	if !strings.Contains(body, lib.artists.Artists[0].Name) {
		t.Fatalf("missing artist %q", lib.artists.Artists[0].Name)
	}
	if !strings.Contains(body, `period=1month`) {
		t.Fatal("period chips should keep 1month")
	}
	if !strings.Contains(body, "listen-share") {
		t.Fatal("missing artist share bar")
	}
}

func TestAlbums_ListsTop(t *testing.T) {
	lib := seedLibrary(t)
	h := newListening(t, lib)
	rr := httptest.NewRecorder()
	h.Albums(rr, httptest.NewRequest(http.MethodGet, "/albuns", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), lib.albums.Albums[0].Name) {
		t.Fatalf("missing album %q", lib.albums.Albums[0].Name)
	}
}

func TestTracks_ListsTop(t *testing.T) {
	lib := seedLibrary(t)
	h := newListening(t, lib)
	rr := httptest.NewRecorder()
	h.Tracks(rr, httptest.NewRequest(http.MethodGet, "/faixas", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), lib.tracks.Tracks[0].Name) {
		t.Fatalf("missing track %q", lib.tracks.Tracks[0].Name)
	}
}

func TestLoved_ListsLovedTracks(t *testing.T) {
	lib := seedLibrary(t)
	h := newListening(t, lib)
	rr := httptest.NewRecorder()
	h.Loved(rr, httptest.NewRequest(http.MethodGet, "/amadas", nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), lib.loved.Tracks[0].Name) {
		t.Fatalf("missing loved %q", lib.loved.Tracks[0].Name)
	}
}

func TestHome_LibraryErrorShowsPortuguese(t *testing.T) {
	lib := seedLibrary(t)
	lib.err = fmt.Errorf("last.fm: User not found")
	h := newListening(t, lib)
	rr := httptest.NewRecorder()
	h.Home(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want 502", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "não deu para falar com o Last.fm") {
		t.Fatalf("missing PT error, body=%s", rr.Body.String())
	}
}

func TestRobots_DisallowsAll(t *testing.T) {
	h := newListening(t, seedLibrary(t))
	rr := httptest.NewRecorder()
	h.Robots(rr, httptest.NewRequest(http.MethodGet, "/robots.txt", nil))
	if got := rr.Body.String(); got != "User-agent: *\nDisallow: /\n" {
		t.Errorf("robots = %q", got)
	}
}
