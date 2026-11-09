package lastfm

import "context"

// Library is the Last.fm read surface the pages need.
type Library interface {
	UserName() string
	User(ctx context.Context) (User, error)
	RecentTracks(ctx context.Context, page, limit int) (RecentPage, error)
	TopArtists(ctx context.Context, period string, page, limit int) (ArtistPage, error)
	TopAlbums(ctx context.Context, period string, page, limit int) (AlbumPage, error)
	TopTracks(ctx context.Context, period string, page, limit int) (TrackPage, error)
	LovedTracks(ctx context.Context, page, limit int) (TrackPage, error)
}
