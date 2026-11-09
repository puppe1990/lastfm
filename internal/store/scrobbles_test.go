package store

import (
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais/fakedata"

	"github.com/puppe1990/lastfm/internal/lastfm"
)

func TestUpsertScrobbles_InsertsSkipsNowPlayingAndDedupes(t *testing.T) {
	fakedata.Seed(7)
	s := newTestStore(t)
	track := fakedata.Title()
	artist := fakedata.Name()
	rows := []lastfm.Track{
		{Name: track, Artist: artist, Album: "obZen", URL: "https://www.last.fm/music/a", Image: "https://img/a.jpg", PlayedUnix: 1_700_000_100},
		{Name: track, Artist: artist, Album: "obZen", URL: "https://www.last.fm/music/a", PlayedUnix: 1_700_000_100},
		{Name: "now", Artist: artist, NowPlaying: true, PlayedUnix: 1_700_000_200},
		{Name: "ghost", Artist: artist, PlayedUnix: 0},
	}

	n, err := s.UpsertScrobbles(rows)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("inserted = %d, want 1", n)
	}

	again, err := s.UpsertScrobbles(rows[:1])
	if err != nil {
		t.Fatal(err)
	}
	if again != 0 {
		t.Fatalf("second insert = %d, want 0", again)
	}

	_, _, count, err := s.ScrobbleBounds()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestListScrobblesBetween_FiltersByUnix(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.UpsertScrobbles([]lastfm.Track{
		{Name: "old", Artist: "A", PlayedUnix: 100},
		{Name: "mid", Artist: "B", PlayedUnix: 200},
		{Name: "new", Artist: "C", PlayedUnix: 300},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := s.ListScrobblesBetween(200, 300)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0].Name != "new" || got[1].Name != "mid" {
		t.Fatalf("order = %q %q", got[0].Name, got[1].Name)
	}
}

func TestTopArtistsBetween_CountsPlays(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.UpsertScrobbles([]lastfm.Track{
		{Name: "Bleed", Artist: "Meshuggah", PlayedUnix: 10},
		{Name: "obZen", Artist: "Meshuggah", PlayedUnix: 11},
		{Name: "Bleed", Artist: "Meshuggah", PlayedUnix: 12},
		{Name: "In Waves", Artist: "Trivium", PlayedUnix: 13},
		{Name: "too-old", Artist: "Opeth", PlayedUnix: 1},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := s.TopArtistsBetween(10, 13, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2: %+v", len(got), got)
	}
	if got[0].Name != "Meshuggah" || got[0].Playcount != 3 || got[0].Rank != 1 {
		t.Fatalf("top = %+v", got[0])
	}
	if got[1].Name != "Trivium" || got[1].Playcount != 1 {
		t.Fatalf("second = %+v", got[1])
	}
}

func TestSyncCursor_RoundTrip(t *testing.T) {
	s := newTestStore(t)
	empty, err := s.LoadSyncCursor()
	if err != nil {
		t.Fatal(err)
	}
	if empty.BackfillDone || empty.NewestUnix != 0 {
		t.Fatalf("empty = %+v", empty)
	}

	want := SyncCursor{NewestUnix: 9, OldestUnix: 2, RegisteredUnix: 1, BackfillDone: true, LastError: "nope"}
	if err := s.SaveSyncCursor(want); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadSyncCursor()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("got %+v want %+v", got, want)
	}
}
