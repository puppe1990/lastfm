package lastfm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const defaultCacheTTL = 45 * time.Second

type cacheEntry struct {
	until time.Time
	val   any
}

// Cache is a short TTL in front of Library so a page of charts does not
// hammer Last.fm on every Drive morph.
type Cache struct {
	inner Library
	ttl   time.Duration
	now   func() time.Time
	mu    sync.Mutex
	items map[string]cacheEntry
}

var _ Library = (*Cache)(nil)

// NewCache wraps inner. ttl <= 0 uses 45s.
func NewCache(inner Library, ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	return &Cache{
		inner: inner,
		ttl:   ttl,
		now:   time.Now,
		items: map[string]cacheEntry{},
	}
}

func (c *Cache) UserName() string {
	return c.inner.UserName()
}

func (c *Cache) User(ctx context.Context) (User, error) {
	val, err := c.fetch("user", func() (any, error) {
		return c.inner.User(ctx)
	})
	if err != nil {
		return User{}, err
	}
	return val.(User), nil
}

func (c *Cache) RecentTracks(ctx context.Context, page, limit int) (RecentPage, error) {
	key := fmt.Sprintf("recent:%d:%d", page, limit)
	val, err := c.fetch(key, func() (any, error) {
		return c.inner.RecentTracks(ctx, page, limit)
	})
	if err != nil {
		return RecentPage{}, err
	}
	return val.(RecentPage), nil
}

func (c *Cache) TopArtists(ctx context.Context, period string, page, limit int) (ArtistPage, error) {
	key := fmt.Sprintf("artists:%s:%d:%d", period, page, limit)
	val, err := c.fetch(key, func() (any, error) {
		return c.inner.TopArtists(ctx, period, page, limit)
	})
	if err != nil {
		return ArtistPage{}, err
	}
	return val.(ArtistPage), nil
}

func (c *Cache) TopAlbums(ctx context.Context, period string, page, limit int) (AlbumPage, error) {
	key := fmt.Sprintf("albums:%s:%d:%d", period, page, limit)
	val, err := c.fetch(key, func() (any, error) {
		return c.inner.TopAlbums(ctx, period, page, limit)
	})
	if err != nil {
		return AlbumPage{}, err
	}
	return val.(AlbumPage), nil
}

func (c *Cache) TopTracks(ctx context.Context, period string, page, limit int) (TrackPage, error) {
	key := fmt.Sprintf("tracks:%s:%d:%d", period, page, limit)
	val, err := c.fetch(key, func() (any, error) {
		return c.inner.TopTracks(ctx, period, page, limit)
	})
	if err != nil {
		return TrackPage{}, err
	}
	return val.(TrackPage), nil
}

func (c *Cache) LovedTracks(ctx context.Context, page, limit int) (TrackPage, error) {
	key := fmt.Sprintf("loved:%d:%d", page, limit)
	val, err := c.fetch(key, func() (any, error) {
		return c.inner.LovedTracks(ctx, page, limit)
	})
	if err != nil {
		return TrackPage{}, err
	}
	return val.(TrackPage), nil
}

func (c *Cache) fetch(key string, load func() (any, error)) (any, error) {
	now := c.now()
	c.mu.Lock()
	if e, ok := c.items[key]; ok && now.Before(e.until) {
		val := e.val
		c.mu.Unlock()
		return val, nil
	}
	c.mu.Unlock()

	val, err := load()
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.items[key] = cacheEntry{until: now.Add(c.ttl), val: val}
	c.mu.Unlock()
	return val, nil
}
