package jobs

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/puppe1990/lastfm/internal/lastfm"
	"github.com/puppe1990/lastfm/internal/store"
)

func TestPerformSyncScrobbles_StoresRangeFromLastfm(t *testing.T) {
	now := time.Now()
	uts := now.Add(-time.Hour).Unix()
	registered := now.Add(-40 * 24 * time.Hour).Unix()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("method") {
		case "user.getInfo":
			fmt.Fprintf(w, `{"user":{"name":"puppe90","playcount":"2","registered":{"unixtime":"%d"}}}`, registered)
		case "user.getRecentTracks":
			from, _ := strconv.ParseInt(r.URL.Query().Get("from"), 10, 64)
			to, _ := strconv.ParseInt(r.URL.Query().Get("to"), 10, 64)
			if (from > 0 && uts <= from) || (to > 0 && uts >= to) {
				fmt.Fprint(w, `{"recenttracks":{"track":[],"@attr":{"page":"1","totalPages":"1","total":"0"}}}`)
				return
			}
			fmt.Fprintf(w, `{"recenttracks":{"track":[{"name":"Bleed","artist":{"name":"Meshuggah"},"date":{"uts":"%d"}}],"@attr":{"page":"1","totalPages":"1","total":"1"}}}`, uts)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	client := lastfm.New(lastfm.Config{APIKey: "k", User: "puppe90", BaseURL: srv.URL}, srv.Client())
	h := PerformSyncScrobbles(s, client)
	if err := h(context.Background(), []byte("{}")); err != nil {
		t.Fatal(err)
	}
	_, _, count, err := s.ScrobbleBounds()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("stored = %d, want 1", count)
	}
	got, err := s.ListScrobblesBetween(0, 2_000_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Artist != "Meshuggah" {
		t.Fatalf("rows = %+v", got)
	}
}
