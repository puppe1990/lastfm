package lastfm

// User is a Last.fm profile (user.getInfo).
type User struct {
	Name           string
	RealName       string
	URL            string
	Country        string
	Image          string
	Playcount      int
	ArtistCount    int
	AlbumCount     int
	TrackCount     int
	RegisteredUnix int64
}

// Track is a scrobble, now-playing row, loved track, or top track.
type Track struct {
	Name       string
	Artist     string
	Album      string
	URL        string
	Image      string
	Playcount  int
	Rank       int
	PlayedUnix int64
	NowPlaying bool
}

// Artist is a top-artist row.
type Artist struct {
	Name      string
	URL       string
	Image     string
	Playcount int
	Rank      int
}

// Album is a top-album row.
type Album struct {
	Name      string
	Artist    string
	URL       string
	Image     string
	Playcount int
	Rank      int
}

// RecentPage is user.getRecentTracks after splitting now-playing from history.
type RecentPage struct {
	NowPlaying *Track
	Tracks     []Track
	Page       int
	TotalPages int
	Total      int
}

// ArtistPage is user.getTopArtists.
type ArtistPage struct {
	Artists    []Artist
	Page       int
	TotalPages int
	Total      int
}

// AlbumPage is user.getTopAlbums.
type AlbumPage struct {
	Albums     []Album
	Page       int
	TotalPages int
	Total      int
}

// TrackPage is user.getTopTracks or user.getLovedTracks.
type TrackPage struct {
	Tracks     []Track
	Page       int
	TotalPages int
	Total      int
}
