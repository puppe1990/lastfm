package lastfm

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type countLib struct {
	n       int
	err     error
	user    User
	recent  RecentPage
	artists ArtistPage
	albums  AlbumPage
	tracks  TrackPage
	loved   TrackPage
}

func (c *countLib) UserName() string { return c.user.Name }

func (c *countLib) User(context.Context) (User, error) {
	c.n++
	if c.err != nil {
		return User{}, c.err
	}
	return c.user, nil
}

func (c *countLib) RecentTracks(context.Context, int, int) (RecentPage, error) {
	c.n++
	if c.err != nil {
		return RecentPage{}, c.err
	}
	return c.recent, nil
}

func (c *countLib) TopArtists(context.Context, string, int, int) (ArtistPage, error) {
	c.n++
	if c.err != nil {
		return ArtistPage{}, c.err
	}
	return c.artists, nil
}

func (c *countLib) TopAlbums(context.Context, string, int, int) (AlbumPage, error) {
	c.n++
	if c.err != nil {
		return AlbumPage{}, c.err
	}
	return c.albums, nil
}

func (c *countLib) TopTracks(context.Context, string, int, int) (TrackPage, error) {
	c.n++
	if c.err != nil {
		return TrackPage{}, c.err
	}
	return c.tracks, nil
}

func (c *countLib) LovedTracks(context.Context, int, int) (TrackPage, error) {
	c.n++
	if c.err != nil {
		return TrackPage{}, c.err
	}
	return c.loved, nil
}

func TestCache_ReusesUserWithinTTL(t *testing.T) {
	inner := &countLib{user: User{Name: "puppe90", Playcount: 3}}
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	c := NewCache(inner, 45*time.Second)
	c.now = func() time.Time { return now }

	first, err := c.User(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	second, err := c.User(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if inner.n != 1 {
		t.Fatalf("inner calls = %d, want 1", inner.n)
	}
	if first.Playcount != 3 || second.Playcount != 3 {
		t.Fatalf("playcount first=%d second=%d", first.Playcount, second.Playcount)
	}

	now = now.Add(46 * time.Second)
	if _, err := c.User(context.Background()); err != nil {
		t.Fatal(err)
	}
	if inner.n != 2 {
		t.Fatalf("after TTL inner calls = %d, want 2", inner.n)
	}
}

func TestCache_DoesNotStoreErrors(t *testing.T) {
	inner := &countLib{err: fmt.Errorf("last.fm: boom")}
	c := NewCache(inner, time.Minute)

	if _, err := c.User(context.Background()); err == nil {
		t.Fatal("expected error")
	}
	inner.err = nil
	inner.user = User{Name: "puppe90"}
	user, err := c.User(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "puppe90" {
		t.Fatalf("user = %+v", user)
	}
	if inner.n != 2 {
		t.Fatalf("inner calls = %d, want 2", inner.n)
	}
}

func TestCache_UserNameDelegates(t *testing.T) {
	c := NewCache(&countLib{user: User{Name: "puppe90"}}, time.Second)
	if got := c.UserName(); got != "puppe90" {
		t.Fatalf("UserName = %q", got)
	}
}
