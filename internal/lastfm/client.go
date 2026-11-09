package lastfm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://ws.audioscrobbler.com/2.0/"

// Config is the Last.fm API key plus the profile we display.
type Config struct {
	APIKey  string
	User    string
	BaseURL string
}

// ConfigFromEnv reads LASTFM_API_KEY (required) and LASTFM_USER (puppe90).
func ConfigFromEnv() (Config, error) {
	key := strings.TrimSpace(os.Getenv("LASTFM_API_KEY"))
	if key == "" {
		return Config{}, fmt.Errorf("LASTFM_API_KEY is required")
	}
	user := strings.TrimSpace(os.Getenv("LASTFM_USER"))
	if user == "" {
		user = "puppe90"
	}
	return Config{APIKey: key, User: user}, nil
}

// Client talks to ws.audioscrobbler.com. HTTP is injected for tests.
type Client struct {
	cfg  Config
	http *http.Client
}

var _ Library = (*Client)(nil)

// New builds a Client. A nil httpClient gets a 12s timeout.
func New(cfg Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 12 * time.Second}
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return &Client{cfg: cfg, http: httpClient}
}

// UserName is the Last.fm user this client reads.
func (c *Client) UserName() string {
	return c.cfg.User
}

// User is user.getInfo.
func (c *Client) User(ctx context.Context) (User, error) {
	raw, err := c.get(ctx, url.Values{"method": {"user.getInfo"}})
	if err != nil {
		return User{}, err
	}
	return ParseUser(raw)
}

// RecentTracks is user.getRecentTracks (extended=1 so artist.name is filled).
func (c *Client) RecentTracks(ctx context.Context, page, limit int) (RecentPage, error) {
	raw, err := c.get(ctx, url.Values{
		"method":   {"user.getRecentTracks"},
		"extended": {"1"},
		"page":     {itoa(page)},
		"limit":    {itoa(limit)},
	})
	if err != nil {
		return RecentPage{}, err
	}
	return ParseRecent(raw)
}

// TopArtists is user.getTopArtists for a Last.fm period.
func (c *Client) TopArtists(ctx context.Context, period string, page, limit int) (ArtistPage, error) {
	raw, err := c.get(ctx, chartQuery("user.getTopArtists", period, page, limit))
	if err != nil {
		return ArtistPage{}, err
	}
	return ParseTopArtists(raw)
}

// TopAlbums is user.getTopAlbums for a Last.fm period.
func (c *Client) TopAlbums(ctx context.Context, period string, page, limit int) (AlbumPage, error) {
	raw, err := c.get(ctx, chartQuery("user.getTopAlbums", period, page, limit))
	if err != nil {
		return AlbumPage{}, err
	}
	return ParseTopAlbums(raw)
}

// TopTracks is user.getTopTracks for a Last.fm period.
func (c *Client) TopTracks(ctx context.Context, period string, page, limit int) (TrackPage, error) {
	raw, err := c.get(ctx, chartQuery("user.getTopTracks", period, page, limit))
	if err != nil {
		return TrackPage{}, err
	}
	return ParseTopTracks(raw)
}

// LovedTracks is user.getLovedTracks.
func (c *Client) LovedTracks(ctx context.Context, page, limit int) (TrackPage, error) {
	raw, err := c.get(ctx, url.Values{
		"method": {"user.getLovedTracks"},
		"page":   {itoa(page)},
		"limit":  {itoa(limit)},
	})
	if err != nil {
		return TrackPage{}, err
	}
	return ParseLoved(raw)
}

func chartQuery(method, period string, page, limit int) url.Values {
	return url.Values{
		"method": {method},
		"period": {NormalizePeriod(period)},
		"page":   {itoa(page)},
		"limit":  {itoa(limit)},
	}
}

func (c *Client) get(ctx context.Context, extra url.Values) ([]byte, error) {
	q := url.Values{
		"method":  extra["method"],
		"user":    {c.cfg.User},
		"api_key": {c.cfg.APIKey},
		"format":  {"json"},
	}
	for k, v := range extra {
		if k == "method" {
			continue
		}
		q[k] = v
	}
	endpoint := strings.TrimRight(c.cfg.BaseURL, "/") + "/?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("last.fm request: %w", err)
	}
	req.Header.Set("User-Agent", "Ouvindo/1.0 (https://lastfm.apps.gestaobem.com)")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("last.fm fetch %s: %w", extra.Get("method"), err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, fmt.Errorf("last.fm body %s: %w", extra.Get("method"), err)
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("last.fm HTTP %d %s", resp.StatusCode, extra.Get("method"))
	}
	return body, nil
}

func itoa(n int) string {
	if n < 1 {
		n = 1
	}
	return fmt.Sprintf("%d", n)
}
