package lastfm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type apiError struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
}

type imageNode struct {
	Size string `json:"size"`
	Text string `json:"#text"`
}

type namedNode struct {
	Name string `json:"name"`
	Text string `json:"#text"`
	URL  string `json:"url"`
}

type attrNode struct {
	NowPlaying string `json:"nowplaying"`
	Rank       string `json:"rank"`
	Page       string `json:"page"`
	TotalPages string `json:"totalPages"`
	Total      string `json:"total"`
}

type dateNode struct {
	UTS      string `json:"uts"`
	UnixTime string `json:"unixtime"`
	Text     any    `json:"#text"`
}

type userEnvelope struct {
	User struct {
		Name        string      `json:"name"`
		RealName    string      `json:"realname"`
		URL         string      `json:"url"`
		Country     string      `json:"country"`
		Playcount   string      `json:"playcount"`
		ArtistCount string      `json:"artist_count"`
		AlbumCount  string      `json:"album_count"`
		TrackCount  string      `json:"track_count"`
		Image       []imageNode `json:"image"`
		Registered  dateNode    `json:"registered"`
	} `json:"user"`
}

type trackNode struct {
	Name   string      `json:"name"`
	URL    string      `json:"url"`
	Artist namedNode   `json:"artist"`
	Album  namedNode   `json:"album"`
	Image  []imageNode `json:"image"`
	Date   dateNode    `json:"date"`
	Attr   attrNode    `json:"@attr"`
	Count  string      `json:"playcount"`
}

type artistNode struct {
	Name  string      `json:"name"`
	URL   string      `json:"url"`
	Image []imageNode `json:"image"`
	Count string      `json:"playcount"`
	Attr  attrNode    `json:"@attr"`
}

type albumNode struct {
	Name   string      `json:"name"`
	URL    string      `json:"url"`
	Artist namedNode   `json:"artist"`
	Image  []imageNode `json:"image"`
	Count  string      `json:"playcount"`
	Attr   attrNode    `json:"@attr"`
}

type recentEnvelope struct {
	Recent struct {
		Track flexList[trackNode] `json:"track"`
		Attr  attrNode            `json:"@attr"`
	} `json:"recenttracks"`
}

type artistsEnvelope struct {
	Top struct {
		Artist flexList[artistNode] `json:"artist"`
		Attr   attrNode             `json:"@attr"`
	} `json:"topartists"`
}

type albumsEnvelope struct {
	Top struct {
		Album flexList[albumNode] `json:"album"`
		Attr  attrNode            `json:"@attr"`
	} `json:"topalbums"`
}

type tracksEnvelope struct {
	Top struct {
		Track flexList[trackNode] `json:"track"`
		Attr  attrNode            `json:"@attr"`
	} `json:"toptracks"`
}

type lovedEnvelope struct {
	Loved struct {
		Track flexList[trackNode] `json:"track"`
		Attr  attrNode            `json:"@attr"`
	} `json:"lovedtracks"`
}

// flexList accepts Last.fm's array-or-single-object JSON.
type flexList[T any] []T

func (l *flexList[T]) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" || s == `""` || s == "[]" {
		*l = nil
		return nil
	}
	if len(s) > 0 && s[0] == '{' {
		var one T
		if err := json.Unmarshal(b, &one); err != nil {
			return err
		}
		*l = []T{one}
		return nil
	}
	var many []T
	if err := json.Unmarshal(b, &many); err != nil {
		return err
	}
	*l = many
	return nil
}

func checkAPIError(raw []byte) error {
	var e apiError
	if err := json.Unmarshal(raw, &e); err != nil {
		return nil
	}
	if e.Error == 0 {
		return nil
	}
	msg := strings.TrimSpace(e.Message)
	if msg == "" {
		msg = fmt.Sprintf("error %d", e.Error)
	}
	return fmt.Errorf("last.fm: %s", msg)
}

func decode[T any](raw []byte) (T, error) {
	var zero T
	if err := checkAPIError(raw); err != nil {
		return zero, err
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, fmt.Errorf("last.fm json: %w", err)
	}
	return out, nil
}

// ParseUser reads user.getInfo JSON.
func ParseUser(raw []byte) (User, error) {
	env, err := decode[userEnvelope](raw)
	if err != nil {
		return User{}, err
	}
	u := env.User
	return User{
		Name:           u.Name,
		RealName:       u.RealName,
		URL:            u.URL,
		Country:        u.Country,
		Image:          pickImage(u.Image),
		Playcount:      atoi(u.Playcount),
		ArtistCount:    atoi(u.ArtistCount),
		AlbumCount:     atoi(u.AlbumCount),
		TrackCount:     atoi(u.TrackCount),
		RegisteredUnix: parseUnix(firstNonEmpty(u.Registered.UTS, u.Registered.UnixTime), u.Registered.Text),
	}, nil
}

// ParseRecent reads user.getRecentTracks JSON.
func ParseRecent(raw []byte) (RecentPage, error) {
	env, err := decode[recentEnvelope](raw)
	if err != nil {
		return RecentPage{}, err
	}
	page := RecentPage{
		Page:       atoi(env.Recent.Attr.Page),
		TotalPages: atoi(env.Recent.Attr.TotalPages),
		Total:      atoi(env.Recent.Attr.Total),
	}
	for _, n := range env.Recent.Track {
		tr := trackFromNode(n)
		if tr.NowPlaying {
			cp := tr
			page.NowPlaying = &cp
			continue
		}
		page.Tracks = append(page.Tracks, tr)
	}
	return page, nil
}

// ParseTopArtists reads user.getTopArtists JSON.
func ParseTopArtists(raw []byte) (ArtistPage, error) {
	env, err := decode[artistsEnvelope](raw)
	if err != nil {
		return ArtistPage{}, err
	}
	out := ArtistPage{
		Page:       atoi(env.Top.Attr.Page),
		TotalPages: atoi(env.Top.Attr.TotalPages),
		Total:      atoi(env.Top.Attr.Total),
	}
	for _, n := range env.Top.Artist {
		out.Artists = append(out.Artists, Artist{
			Name:      n.Name,
			URL:       n.URL,
			Image:     pickImage(n.Image),
			Playcount: atoi(n.Count),
			Rank:      atoi(n.Attr.Rank),
		})
	}
	return out, nil
}

// ParseTopAlbums reads user.getTopAlbums JSON.
func ParseTopAlbums(raw []byte) (AlbumPage, error) {
	env, err := decode[albumsEnvelope](raw)
	if err != nil {
		return AlbumPage{}, err
	}
	out := AlbumPage{
		Page:       atoi(env.Top.Attr.Page),
		TotalPages: atoi(env.Top.Attr.TotalPages),
		Total:      atoi(env.Top.Attr.Total),
	}
	for _, n := range env.Top.Album {
		out.Albums = append(out.Albums, Album{
			Name:      n.Name,
			Artist:    firstNonEmpty(n.Artist.Name, n.Artist.Text),
			URL:       n.URL,
			Image:     pickImage(n.Image),
			Playcount: atoi(n.Count),
			Rank:      atoi(n.Attr.Rank),
		})
	}
	return out, nil
}

// ParseTopTracks reads user.getTopTracks JSON.
func ParseTopTracks(raw []byte) (TrackPage, error) {
	env, err := decode[tracksEnvelope](raw)
	if err != nil {
		return TrackPage{}, err
	}
	return TrackPage{
		Tracks:     tracksFrom(env.Top.Track),
		Page:       atoi(env.Top.Attr.Page),
		TotalPages: atoi(env.Top.Attr.TotalPages),
		Total:      atoi(env.Top.Attr.Total),
	}, nil
}

// ParseLoved reads user.getLovedTracks JSON.
func ParseLoved(raw []byte) (TrackPage, error) {
	env, err := decode[lovedEnvelope](raw)
	if err != nil {
		return TrackPage{}, err
	}
	return TrackPage{
		Tracks:     tracksFrom(env.Loved.Track),
		Page:       atoi(env.Loved.Attr.Page),
		TotalPages: atoi(env.Loved.Attr.TotalPages),
		Total:      atoi(env.Loved.Attr.Total),
	}, nil
}

func tracksFrom(nodes flexList[trackNode]) []Track {
	out := make([]Track, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, trackFromNode(n))
	}
	return out
}

func trackFromNode(n trackNode) Track {
	return Track{
		Name:       n.Name,
		Artist:     firstNonEmpty(n.Artist.Name, n.Artist.Text),
		Album:      firstNonEmpty(n.Album.Text, n.Album.Name),
		URL:        n.URL,
		Image:      pickImage(n.Image),
		Playcount:  atoi(n.Count),
		Rank:       atoi(n.Attr.Rank),
		PlayedUnix: parseUnix(firstNonEmpty(n.Date.UTS, n.Date.UnixTime), n.Date.Text),
		NowPlaying: n.Attr.NowPlaying == "true",
	}
}

func pickImage(images []imageNode) string {
	bySize := map[string]string{}
	for _, img := range images {
		if img.Text == "" {
			continue
		}
		bySize[img.Size] = img.Text
	}
	for _, size := range []string{"extralarge", "large", "medium", "mega", "small"} {
		if u := bySize[size]; u != "" {
			return u
		}
	}
	for _, img := range images {
		if img.Text != "" {
			return img.Text
		}
	}
	return ""
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

func parseUnix(uts string, text any) int64 {
	if n, err := strconv.ParseInt(strings.TrimSpace(uts), 10, 64); err == nil && n > 0 {
		return n
	}
	switch v := text.(type) {
	case float64:
		return int64(v)
	case json.Number:
		n, _ := v.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		return n
	}
	return 0
}
