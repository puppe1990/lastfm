package store

import (
	"testing"
	"time"

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

func TestTopAlbumsByYear_RanksByPlaysAndSkipsBlank(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	s := newTestStore(t)
	y2024 := time.Date(2024, 6, 1, 12, 0, 0, 0, loc).Unix()
	y2023 := time.Date(2023, 6, 1, 12, 0, 0, 0, loc).Unix()
	nyeBRT := time.Date(2023, 12, 31, 23, 30, 0, 0, loc).Unix()
	if _, err := s.UpsertScrobbles([]lastfm.Track{
		{Name: "Bleed", Artist: "Meshuggah", Album: "obZen", Image: "https://img/obzen.jpg", PlayedUnix: y2024},
		{Name: "Bleed", Artist: "Meshuggah", Album: "obZen", PlayedUnix: y2024 + 60},
		{Name: "Bleed", Artist: "Meshuggah", Album: "obZen", PlayedUnix: y2024 + 120},
		{Name: "Dancers", Artist: "Meshuggah", Album: "Catch Thirtythree", PlayedUnix: y2024 + 180},
		{Name: "Ghost", Artist: "Opeth", Album: "Blackwater Park", PlayedUnix: y2023},
		{Name: "Ghost", Artist: "Opeth", Album: "Blackwater Park", PlayedUnix: y2023 + 60},
		{Name: "nye", Artist: "Meshuggah", Album: "obZen", PlayedUnix: nyeBRT},
		{Name: "loose", Artist: "X", Album: "", PlayedUnix: y2024 + 200},
		{Name: "spaces", Artist: "Y", Album: "   ", PlayedUnix: y2024 + 240},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := s.TopAlbumsByYear(loc, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("years = %d, want 2: %+v", len(got), got)
	}
	if got[0].Year != 2024 || got[1].Year != 2023 {
		t.Fatalf("year order = %d %d", got[0].Year, got[1].Year)
	}
	if len(got[0].Albums) != 2 {
		t.Fatalf("2024 albums = %+v", got[0].Albums)
	}
	top := got[0].Albums[0]
	if top.Name != "obZen" || top.Artist != "Meshuggah" || top.Playcount != 3 || top.Rank != 1 {
		t.Fatalf("2024 #1 = %+v", top)
	}
	if top.URL != "https://www.last.fm/music/Meshuggah/obZen" {
		t.Errorf("url = %q", top.URL)
	}
	if got[0].Albums[1].Name != "Catch Thirtythree" || got[0].Albums[1].Playcount != 1 {
		t.Fatalf("2024 #2 = %+v", got[0].Albums[1])
	}
	if len(got[1].Albums) != 2 {
		t.Fatalf("2023 albums = %+v", got[1].Albums)
	}
	if got[1].Albums[0].Name != "Blackwater Park" || got[1].Albums[0].Playcount != 2 {
		t.Fatalf("2023 #1 = %+v", got[1].Albums[0])
	}
	if got[1].Albums[1].Name != "obZen" || got[1].Albums[1].Playcount != 1 {
		t.Fatalf("NYE BRT should stay in 2023: %+v", got[1].Albums[1])
	}

	limited, err := s.TopAlbumsByYear(loc, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(limited[0].Albums) != 1 || limited[0].Albums[0].Name != "obZen" {
		t.Fatalf("limit 1 = %+v", limited[0].Albums)
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
