package lastfm

import (
	"fmt"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais/fakedata"
)

func TestParseUser_ReadsPlaycountNameAndCover(t *testing.T) {
	fakedata.Seed(42)
	realName := fakedata.Name()
	cover := fakedata.URL() + "/cover.png"
	raw := fmt.Sprintf(`{
		"user": {
			"name": "puppe90",
			"realname": %q,
			"playcount": "82197",
			"artist_count": "2957",
			"album_count": "9017",
			"track_count": "23622",
			"url": "https://www.last.fm/user/puppe90",
			"country": "Brazil",
			"image": [
				{"size": "small", "#text": "https://img.example/small.png"},
				{"size": "extralarge", "#text": %q}
			],
			"registered": {"unixtime": "1201408060"}
		}
	}`, realName, cover)

	user, err := ParseUser([]byte(raw))
	if err != nil {
		t.Fatalf("ParseUser: %v", err)
	}
	if user.Name != "puppe90" {
		t.Errorf("Name = %q", user.Name)
	}
	if user.RealName != realName {
		t.Errorf("RealName = %q, want %q", user.RealName, realName)
	}
	if user.Playcount != 82197 {
		t.Errorf("Playcount = %d", user.Playcount)
	}
	if user.ArtistCount != 2957 || user.AlbumCount != 9017 || user.TrackCount != 23622 {
		t.Errorf("counts artists=%d albums=%d tracks=%d", user.ArtistCount, user.AlbumCount, user.TrackCount)
	}
	if user.Image != cover {
		t.Errorf("Image = %q, want extralarge %q", user.Image, cover)
	}
	if user.RegisteredUnix != 1201408060 {
		t.Errorf("RegisteredUnix = %d", user.RegisteredUnix)
	}
	if user.URL != "https://www.last.fm/user/puppe90" {
		t.Errorf("URL = %q", user.URL)
	}
}

func TestParseRecent_SplitsNowPlayingFromHistory(t *testing.T) {
	fakedata.Seed(7)
	playing := fakedata.Title()
	past := fakedata.Title()
	artist := fakedata.Name()
	album := fakedata.Title()
	raw := fmt.Sprintf(`{
		"recenttracks": {
			"track": [
				{
					"name": %q,
					"artist": {"name": %q},
					"album": {"#text": %q},
					"url": "https://www.last.fm/music/x",
					"image": [{"size": "large", "#text": "https://img.example/now.jpg"}],
					"@attr": {"nowplaying": "true"}
				},
				{
					"name": %q,
					"artist": {"name": %q},
					"album": {"#text": %q},
					"url": "https://www.last.fm/music/y",
					"date": {"uts": "1794246521", "#text": "09 Nov 2026, 17:48"},
					"image": [{"size": "large", "#text": "https://img.example/past.jpg"}]
				}
			],
			"@attr": {"user": "puppe90", "page": "1", "totalPages": "41099", "total": "82197", "perPage": "2"}
		}
	}`, playing, artist, album, past, artist, album)

	page, err := ParseRecent([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRecent: %v", err)
	}
	if page.NowPlaying == nil {
		t.Fatal("expected now playing track")
	}
	if page.NowPlaying.Name != playing {
		t.Errorf("now playing name = %q", page.NowPlaying.Name)
	}
	if page.NowPlaying.Artist != artist {
		t.Errorf("now playing artist = %q", page.NowPlaying.Artist)
	}
	if !page.NowPlaying.NowPlaying {
		t.Error("NowPlaying flag false")
	}
	if len(page.Tracks) != 1 {
		t.Fatalf("history len = %d, want 1 (now playing excluded)", len(page.Tracks))
	}
	if page.Tracks[0].Name != past {
		t.Errorf("history name = %q", page.Tracks[0].Name)
	}
	if page.Tracks[0].PlayedUnix != 1794246521 {
		t.Errorf("PlayedUnix = %d", page.Tracks[0].PlayedUnix)
	}
	if page.Total != 82197 {
		t.Errorf("Total = %d", page.Total)
	}
	if page.Page != 1 || page.TotalPages != 41099 {
		t.Errorf("page %d/%d", page.Page, page.TotalPages)
	}
}

func TestParseRecent_SingleTrackObject(t *testing.T) {
	fakedata.Seed(11)
	name := fakedata.Title()
	raw := fmt.Sprintf(`{
		"recenttracks": {
			"track": {
				"name": %q,
				"artist": {"#text": "Meshuggah"},
				"album": {"#text": "obZen"},
				"date": {"uts": "100"}
			},
			"@attr": {"page": "1", "totalPages": "1", "total": "1"}
		}
	}`, name)
	page, err := ParseRecent([]byte(raw))
	if err != nil {
		t.Fatalf("ParseRecent: %v", err)
	}
	if len(page.Tracks) != 1 || page.Tracks[0].Name != name {
		t.Fatalf("tracks = %+v", page.Tracks)
	}
	if page.Tracks[0].Artist != "Meshuggah" {
		t.Errorf("artist from #text = %q", page.Tracks[0].Artist)
	}
}

func TestParseTopArtists_ReadsRankAndPlaycount(t *testing.T) {
	fakedata.Seed(3)
	name := fakedata.Name()
	raw := fmt.Sprintf(`{
		"topartists": {
			"artist": [{"name": %q, "playcount": "12", "url": "https://www.last.fm/music/x", "@attr": {"rank": "1"}}],
			"@attr": {"page": "1", "totalPages": "4", "total": "8"}
		}
	}`, name)
	page, err := ParseTopArtists([]byte(raw))
	if err != nil {
		t.Fatalf("ParseTopArtists: %v", err)
	}
	if len(page.Artists) != 1 {
		t.Fatalf("len = %d", len(page.Artists))
	}
	if page.Artists[0].Name != name || page.Artists[0].Playcount != 12 || page.Artists[0].Rank != 1 {
		t.Errorf("artist = %+v", page.Artists[0])
	}
}

func TestParseTopAlbums_ReadsArtistName(t *testing.T) {
	fakedata.Seed(5)
	album := fakedata.Title()
	artist := fakedata.Name()
	raw := fmt.Sprintf(`{
		"topalbums": {
			"album": [{"name": %q, "playcount": "12", "artist": {"name": %q}, "image": [{"size": "extralarge", "#text": "https://img.example/a.jpg"}]}],
			"@attr": {"page": "1", "totalPages": "1", "total": "1"}
		}
	}`, album, artist)
	page, err := ParseTopAlbums([]byte(raw))
	if err != nil {
		t.Fatalf("ParseTopAlbums: %v", err)
	}
	if len(page.Albums) != 1 {
		t.Fatalf("len = %d", len(page.Albums))
	}
	if page.Albums[0].Name != album || page.Albums[0].Artist != artist {
		t.Errorf("album = %+v", page.Albums[0])
	}
}

func TestParseAPIError(t *testing.T) {
	_, err := ParseUser([]byte(`{"error":6,"message":"User not found"}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if got := err.Error(); got != "last.fm: User not found" {
		t.Errorf("err = %q", got)
	}
}

func TestNormalizePeriod_FallsBackToSevenDay(t *testing.T) {
	if got := NormalizePeriod("1month"); got != "1month" {
		t.Errorf("1month = %q", got)
	}
	if got := NormalizePeriod("weird"); got != "7day" {
		t.Errorf("unknown = %q", got)
	}
	if got := NormalizePeriod(""); got != "7day" {
		t.Errorf("empty = %q", got)
	}
}

func TestParseJSON_RejectsGarbage(t *testing.T) {
	if _, err := ParseUser([]byte(`not-json`)); err == nil {
		t.Fatal("expected json error")
	}
}
