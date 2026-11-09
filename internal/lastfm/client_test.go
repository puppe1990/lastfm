package lastfm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais/fakedata"
)

func TestConfigFromEnv_RequiresAPIKey(t *testing.T) {
	t.Setenv("LASTFM_API_KEY", "")
	t.Setenv("LASTFM_USER", "")
	if _, err := ConfigFromEnv(); err == nil {
		t.Fatal("expected missing key error")
	}
}

func TestConfigFromEnv_DefaultsUserToPuppe90(t *testing.T) {
	t.Setenv("LASTFM_API_KEY", "abc123")
	t.Setenv("LASTFM_USER", "")
	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "abc123" {
		t.Errorf("APIKey = %q", cfg.APIKey)
	}
	if cfg.User != "puppe90" {
		t.Errorf("User = %q, want puppe90", cfg.User)
	}
}

func TestClient_UserSendsMethodAndKey(t *testing.T) {
	fakedata.Seed(21)
	realName := fakedata.Name()
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		fmt.Fprintf(w, `{"user":{"name":"puppe90","realname":%q,"playcount":"3","registered":{"unixtime":"1"}}}`, realName)
	}))
	t.Cleanup(srv.Close)

	c := New(Config{APIKey: "k3y", User: "puppe90", BaseURL: srv.URL}, srv.Client())
	user, err := c.User(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.RealName != realName {
		t.Errorf("RealName = %q", user.RealName)
	}
	if got, want := gotURL, "/?api_key=k3y&format=json&method=user.getInfo&user=puppe90"; got != want {
		t.Errorf("url = %q, want %q", got, want)
	}
}

func TestClient_RecentTracksUsesExtendedFlag(t *testing.T) {
	fakedata.Seed(22)
	title := fakedata.Title()
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		fmt.Fprintf(w, `{"recenttracks":{"track":[{"name":%q,"artist":{"name":"A"},"date":{"uts":"9"}}],"@attr":{"page":"2","totalPages":"5","total":"10"}}}`, title)
	}))
	t.Cleanup(srv.Close)

	c := New(Config{APIKey: "k", User: "puppe90", BaseURL: srv.URL}, srv.Client())
	page, err := c.RecentTracks(context.Background(), 2, 20)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Tracks) != 1 || page.Tracks[0].Name != title {
		t.Fatalf("tracks = %+v", page.Tracks)
	}
	if page.Page != 2 {
		t.Errorf("page = %d", page.Page)
	}
	q := mustQuery(t, gotQuery)
	if q.Get("method") != "user.getRecentTracks" {
		t.Errorf("method = %q", q.Get("method"))
	}
	if q.Get("extended") != "1" {
		t.Errorf("extended = %q", q.Get("extended"))
	}
	if q.Get("page") != "2" || q.Get("limit") != "20" {
		t.Errorf("page/limit = %s/%s", q.Get("page"), q.Get("limit"))
	}
}

func TestClient_TopArtistsSendsPeriod(t *testing.T) {
	fakedata.Seed(23)
	name := fakedata.Name()
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		fmt.Fprintf(w, `{"topartists":{"artist":[{"name":%q,"playcount":"4"}],"@attr":{"page":"1","total":"1"}}}`, name)
	}))
	t.Cleanup(srv.Close)

	c := New(Config{APIKey: "k", User: "puppe90", BaseURL: srv.URL}, srv.Client())
	page, err := c.TopArtists(context.Background(), "1month", 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Artists) != 1 || page.Artists[0].Name != name {
		t.Fatalf("artists = %+v", page.Artists)
	}
	q := mustQuery(t, gotQuery)
	if q.Get("method") != "user.getTopArtists" {
		t.Errorf("method = %q", q.Get("method"))
	}
	if q.Get("period") != "1month" {
		t.Errorf("period = %q", q.Get("period"))
	}
}

func TestClient_SurfacesLastfmError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"error":6,"message":"User not found"}`)
	}))
	t.Cleanup(srv.Close)
	c := New(Config{APIKey: "k", User: "nope", BaseURL: srv.URL}, srv.Client())
	if _, err := c.User(context.Background()); err == nil || err.Error() != "last.fm: User not found" {
		t.Fatalf("err = %v", err)
	}
}

func TestClient_HTTPStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		io.WriteString(w, "nope")
	}))
	t.Cleanup(srv.Close)
	c := New(Config{APIKey: "k", User: "puppe90", BaseURL: srv.URL}, srv.Client())
	if _, err := c.User(context.Background()); err == nil {
		t.Fatal("expected HTTP error")
	}
}

func TestConfigFromEnv_HonorsUser(t *testing.T) {
	t.Setenv("LASTFM_API_KEY", "k")
	t.Setenv("LASTFM_USER", "other")
	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.User != "other" {
		t.Errorf("User = %q", cfg.User)
	}
}

func mustQuery(t *testing.T, raw string) url.Values {
	t.Helper()
	q, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	return q
}
