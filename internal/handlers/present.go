package handlers

import (
	"time"

	"github.com/puppe1990/lastfm/internal/lastfm"
)

type listenPage struct {
	User        lastfm.User
	Playcount   string
	ArtistCount string
	AlbumCount  string
	TrackCount  string
	SinceYear   string
	NowPlaying  *trackView
	Recent      []trackView
	Artists     []artistView
	Albums      []albumView
	Tracks      []trackView
	Period      string
	Periods     []periodChip
	ShowPeriod  bool
	PageN       int
	TotalPages  int
	PrevPage    int
	NextPage    int
	Path        string
	Error       string
}

type trackView struct {
	Name       string
	Artist     string
	Album      string
	URL        string
	Image      string
	When       string
	Playcount  string
	NowPlaying bool
}

type artistView struct {
	Name      string
	URL       string
	Image     string
	Playcount string
	Rank      int
}

type albumView struct {
	Name      string
	Artist    string
	URL       string
	Image     string
	Playcount string
	Rank      int
}

type periodChip struct {
	ID     string
	Label  string
	Active bool
}

func presentHome(user lastfm.User, recent lastfm.RecentPage, pageN int, now time.Time) listenPage {
	page := listenPage{
		User:        user,
		Playcount:   lastfm.FormatPlaycount(user.Playcount),
		ArtistCount: lastfm.FormatPlaycount(user.ArtistCount),
		AlbumCount:  lastfm.FormatPlaycount(user.AlbumCount),
		TrackCount:  lastfm.FormatPlaycount(user.TrackCount),
		SinceYear:   lastfm.FormatRegisteredYear(user.RegisteredUnix),
		Recent:      presentTrackList(recent.Tracks, now),
		Path:        "/",
		PageN:       pageN,
		TotalPages:  recent.TotalPages,
	}
	if recent.NowPlaying != nil {
		tv := presentTrack(*recent.NowPlaying, now)
		page.NowPlaying = &tv
	}
	fillPager(&page)
	return page
}

func presentArtists(chart lastfm.ArtistPage, period string, pageN int) listenPage {
	items := make([]artistView, 0, len(chart.Artists))
	for _, a := range chart.Artists {
		items = append(items, artistView{
			Name:      a.Name,
			URL:       a.URL,
			Image:     a.Image,
			Playcount: lastfm.FormatPlaycount(a.Playcount),
			Rank:      a.Rank,
		})
	}
	page := listenPage{
		Artists:    items,
		Period:     period,
		Periods:    periodChips(period),
		ShowPeriod: true,
		Path:       "/artistas",
		PageN:      pageN,
		TotalPages: chart.TotalPages,
	}
	fillPager(&page)
	return page
}

func presentAlbums(chart lastfm.AlbumPage, period string, pageN int) listenPage {
	items := make([]albumView, 0, len(chart.Albums))
	for _, a := range chart.Albums {
		items = append(items, albumView{
			Name:      a.Name,
			Artist:    a.Artist,
			URL:       a.URL,
			Image:     a.Image,
			Playcount: lastfm.FormatPlaycount(a.Playcount),
			Rank:      a.Rank,
		})
	}
	page := listenPage{
		Albums:     items,
		Period:     period,
		Periods:    periodChips(period),
		ShowPeriod: true,
		Path:       "/albuns",
		PageN:      pageN,
		TotalPages: chart.TotalPages,
	}
	fillPager(&page)
	return page
}

func presentTracks(chart lastfm.TrackPage, period string, pageN int) listenPage {
	page := listenPage{
		Tracks:     presentTrackList(chart.Tracks, time.Now()),
		Period:     period,
		Periods:    periodChips(period),
		ShowPeriod: true,
		Path:       "/faixas",
		PageN:      pageN,
		TotalPages: chart.TotalPages,
	}
	fillPager(&page)
	return page
}

func presentLoved(chart lastfm.TrackPage, pageN int) listenPage {
	page := listenPage{
		Tracks:     presentTrackList(chart.Tracks, time.Now()),
		Path:       "/amadas",
		PageN:      pageN,
		TotalPages: chart.TotalPages,
	}
	fillPager(&page)
	return page
}

func presentTrackList(tracks []lastfm.Track, now time.Time) []trackView {
	out := make([]trackView, 0, len(tracks))
	for _, tr := range tracks {
		out = append(out, presentTrack(tr, now))
	}
	return out
}

func presentTrack(tr lastfm.Track, now time.Time) trackView {
	return trackView{
		Name:       tr.Name,
		Artist:     tr.Artist,
		Album:      tr.Album,
		URL:        tr.URL,
		Image:      tr.Image,
		When:       lastfm.FormatWhen(tr.PlayedUnix, now),
		Playcount:  lastfm.FormatPlaycount(tr.Playcount),
		NowPlaying: tr.NowPlaying,
	}
}

func periodChips(active string) []periodChip {
	labels := map[string]string{
		"7day":    "7 dias",
		"1month":  "1 mês",
		"3month":  "3 meses",
		"6month":  "6 meses",
		"12month": "12 meses",
		"overall": "desde sempre",
	}
	out := make([]periodChip, 0, len(lastfm.Periods))
	for _, id := range lastfm.Periods {
		out = append(out, periodChip{ID: id, Label: labels[id], Active: id == active})
	}
	return out
}

func fillPager(page *listenPage) {
	if page.PageN < 1 {
		page.PageN = 1
	}
	if page.PageN > 1 {
		page.PrevPage = page.PageN - 1
	}
	if page.TotalPages > page.PageN {
		page.NextPage = page.PageN + 1
	}
}
