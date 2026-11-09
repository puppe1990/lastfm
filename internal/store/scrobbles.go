package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/puppe1990/lastfm/internal/lastfm"
)

// SyncCursor is the Last.fm catch-up watermark for one profile.
type SyncCursor struct {
	NewestUnix     int64
	OldestUnix     int64
	RegisteredUnix int64
	BackfillDone   bool
	LastError      string
}

func (s *SQLiteStore) UpsertScrobbles(tracks []lastfm.Track) (int, error) {
	tx, err := s.db.Raw().Begin()
	if err != nil {
		return 0, fmt.Errorf("begin scrobbles: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`
INSERT OR IGNORE INTO scrobbles (played_unix, track, artist, album, url, image)
VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, fmt.Errorf("prepare scrobbles: %w", err)
	}
	defer stmt.Close()

	var n int
	for _, tr := range tracks {
		if tr.NowPlaying || tr.PlayedUnix <= 0 || tr.Name == "" || tr.Artist == "" {
			continue
		}
		res, err := stmt.Exec(tr.PlayedUnix, tr.Name, tr.Artist, tr.Album, tr.URL, tr.Image)
		if err != nil {
			return 0, fmt.Errorf("insert scrobble %q %q at %d: %w", tr.Artist, tr.Name, tr.PlayedUnix, err)
		}
		got, _ := res.RowsAffected()
		n += int(got)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit scrobbles: %w", err)
	}
	return n, nil
}

func (s *SQLiteStore) ListScrobblesBetween(fromUnix, toUnix int64) ([]lastfm.Track, error) {
	rows, err := s.db.Raw().Query(`
SELECT played_unix, track, artist, album, url, image
FROM scrobbles
WHERE played_unix >= ? AND played_unix <= ?
ORDER BY played_unix DESC`, fromUnix, toUnix)
	if err != nil {
		return nil, fmt.Errorf("list scrobbles %d..%d: %w", fromUnix, toUnix, err)
	}
	defer rows.Close()

	var out []lastfm.Track
	for rows.Next() {
		var tr lastfm.Track
		if err := rows.Scan(&tr.PlayedUnix, &tr.Name, &tr.Artist, &tr.Album, &tr.URL, &tr.Image); err != nil {
			return nil, fmt.Errorf("scan scrobble: %w", err)
		}
		out = append(out, tr)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) TopArtistsBetween(fromUnix, toUnix int64, limit int) ([]lastfm.Artist, error) {
	if limit < 1 {
		limit = 8
	}
	rows, err := s.db.Raw().Query(`
SELECT artist, COUNT(*) AS n
FROM scrobbles
WHERE played_unix >= ? AND played_unix <= ?
GROUP BY artist
ORDER BY n DESC, artist ASC
LIMIT ?`, fromUnix, toUnix, limit)
	if err != nil {
		return nil, fmt.Errorf("top artists %d..%d: %w", fromUnix, toUnix, err)
	}
	defer rows.Close()

	var out []lastfm.Artist
	rank := 1
	for rows.Next() {
		var a lastfm.Artist
		if err := rows.Scan(&a.Name, &a.Playcount); err != nil {
			return nil, fmt.Errorf("scan top artist: %w", err)
		}
		a.Rank = rank
		a.URL = "https://www.last.fm/music/" + lastfmArtistPath(a.Name)
		rank++
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *SQLiteStore) ScrobbleBounds() (minUnix, maxUnix int64, count int, err error) {
	err = s.db.Raw().QueryRow(`
SELECT COALESCE(MIN(played_unix), 0), COALESCE(MAX(played_unix), 0), COUNT(*)
FROM scrobbles`).Scan(&minUnix, &maxUnix, &count)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("scrobble bounds: %w", err)
	}
	return minUnix, maxUnix, count, nil
}

func (s *SQLiteStore) LoadSyncCursor() (SyncCursor, error) {
	var c SyncCursor
	var done int
	err := s.db.Raw().QueryRow(`
SELECT newest_unix, oldest_unix, registered_unix, backfill_done, last_error
FROM scrobble_sync WHERE id = 1`).Scan(&c.NewestUnix, &c.OldestUnix, &c.RegisteredUnix, &done, &c.LastError)
	if err == sql.ErrNoRows {
		return SyncCursor{}, nil
	}
	if err != nil {
		return SyncCursor{}, fmt.Errorf("load sync cursor: %w", err)
	}
	c.BackfillDone = done == 1
	return c, nil
}

func (s *SQLiteStore) SaveSyncCursor(c SyncCursor) error {
	done := 0
	if c.BackfillDone {
		done = 1
	}
	_, err := s.db.Raw().Exec(`
INSERT INTO scrobble_sync (id, newest_unix, oldest_unix, registered_unix, backfill_done, last_error, updated_at)
VALUES (1, ?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
  newest_unix = excluded.newest_unix,
  oldest_unix = excluded.oldest_unix,
  registered_unix = excluded.registered_unix,
  backfill_done = excluded.backfill_done,
  last_error = excluded.last_error,
  updated_at = excluded.updated_at`,
		c.NewestUnix, c.OldestUnix, c.RegisteredUnix, done, c.LastError, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("save sync cursor: %w", err)
	}
	return nil
}

func lastfmArtistPath(name string) string {
	out := make([]rune, 0, len(name))
	for _, r := range name {
		if r == ' ' {
			out = append(out, '+')
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
