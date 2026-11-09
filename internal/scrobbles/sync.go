package scrobbles

import (
	"context"
	"fmt"
	"time"

	"github.com/puppe1990/lastfm/internal/lastfm"
	"github.com/puppe1990/lastfm/internal/store"
)

const (
	defaultPageSize = 200
	defaultGap      = 200 * time.Millisecond
	dayOverlap      = 25 * time.Hour
	backfillWindow  = 30 * 24 * time.Hour
	// DefaultBackfillBudget is how long one job spends walking older than last day.
	DefaultBackfillBudget = 3 * time.Minute
)

// Source is the Last.fm windowed recent-tracks API.
type Source interface {
	User(ctx context.Context) (lastfm.User, error)
	RecentTracksRange(ctx context.Context, fromUnix, toUnix int64, page, limit int) (lastfm.RecentPage, error)
}

// Ledger persists scrobbles and the catch-up cursor.
type Ledger interface {
	UpsertScrobbles(tracks []lastfm.Track) (int, error)
	ScrobbleBounds() (minUnix, maxUnix int64, count int, err error)
	LoadSyncCursor() (store.SyncCursor, error)
	SaveSyncCursor(store.SyncCursor) error
}

// Syncer copies Last.fm history into the ledger: last day first, then older pages.
type Syncer struct {
	Source         Source
	Ledger         Ledger
	Now            func() time.Time
	Sleep          func(time.Duration)
	PageSize       int
	BackfillBudget time.Duration
	PageGap        time.Duration
}

// Report is one sync pass.
type Report struct {
	Inserted     int
	Stored       int
	BackfillDone bool
}

func (s *Syncer) defaults() {
	if s.Now == nil {
		s.Now = time.Now
	}
	if s.Sleep == nil {
		s.Sleep = time.Sleep
	}
	if s.PageSize < 1 {
		s.PageSize = defaultPageSize
	}
	if s.PageGap <= 0 {
		s.PageGap = defaultGap
	}
}

// Run fetches the last 25h, then walks older scrobbles until registered (or the budget).
func (s *Syncer) Run(ctx context.Context) (Report, error) {
	s.defaults()
	now := s.Now()
	user, err := s.Source.User(ctx)
	if err != nil {
		return Report{}, fmt.Errorf("last.fm user: %w", err)
	}
	cursor, err := s.Ledger.LoadSyncCursor()
	if err != nil {
		return Report{}, err
	}
	cursor.RegisteredUnix = user.RegisteredUnix
	cursor.LastError = ""

	from := now.Add(-dayOverlap).Unix()
	if cursor.RegisteredUnix > 0 && from < cursor.RegisteredUnix {
		from = cursor.RegisteredUnix
	}
	inserted, err := s.pullPages(ctx, from, 0)
	if err != nil {
		cursor.LastError = err.Error()
		_ = s.Ledger.SaveSyncCursor(cursor)
		return Report{}, err
	}

	if !cursor.BackfillDone {
		n, err := s.backfill(ctx, &cursor, now)
		inserted += n
		if err != nil {
			cursor.LastError = err.Error()
			_ = s.Ledger.SaveSyncCursor(cursor)
			return Report{}, err
		}
	}

	minUnix, maxUnix, count, err := s.Ledger.ScrobbleBounds()
	if err != nil {
		return Report{}, err
	}
	if maxUnix > cursor.NewestUnix {
		cursor.NewestUnix = maxUnix
	}
	if minUnix > 0 && (cursor.OldestUnix == 0 || minUnix < cursor.OldestUnix) {
		cursor.OldestUnix = minUnix
	}
	if err := s.Ledger.SaveSyncCursor(cursor); err != nil {
		return Report{}, err
	}
	return Report{Inserted: inserted, Stored: count, BackfillDone: cursor.BackfillDone}, nil
}

func (s *Syncer) pullPages(ctx context.Context, fromUnix, toUnix int64) (int, error) {
	inserted := 0
	for page := 1; ; page++ {
		if err := ctx.Err(); err != nil {
			return inserted, err
		}
		got, err := s.Source.RecentTracksRange(ctx, fromUnix, toUnix, page, s.PageSize)
		if err != nil {
			return inserted, fmt.Errorf("recent %d..%d page %d: %w", fromUnix, toUnix, page, err)
		}
		n, err := s.Ledger.UpsertScrobbles(got.Tracks)
		if err != nil {
			return inserted, err
		}
		inserted += n
		if len(got.Tracks) == 0 || page >= got.TotalPages {
			return inserted, nil
		}
		s.Sleep(s.PageGap)
	}
}

func (s *Syncer) backfill(ctx context.Context, cursor *store.SyncCursor, now time.Time) (int, error) {
	if s.BackfillBudget <= 0 {
		return 0, nil
	}
	deadline := now.Add(s.BackfillBudget)
	toUnix, err := s.backfillTo(*cursor)
	if err != nil {
		return 0, err
	}
	inserted := 0
	for s.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return inserted, err
		}
		n, done, err := s.backfillWindow(ctx, cursor, toUnix)
		inserted += n
		if err != nil || done {
			return inserted, err
		}
		toUnix = cursor.OldestUnix
		s.Sleep(s.PageGap)
	}
	return inserted, nil
}

// Last.fm 500s on multi-year from/to; each pass covers at most 30 days.
func (s *Syncer) backfillWindow(ctx context.Context, cursor *store.SyncCursor, toUnix int64) (int, bool, error) {
	fromUnix := windowFrom(toUnix, cursor.RegisteredUnix)
	if toUnix > 0 && fromUnix >= toUnix {
		cursor.BackfillDone = true
		return 0, true, s.Ledger.SaveSyncCursor(*cursor)
	}
	n, err := s.pullPages(ctx, fromUnix, toUnix)
	if err != nil {
		return 0, false, fmt.Errorf("backfill %d..%d: %w", fromUnix, toUnix, err)
	}
	cursor.OldestUnix = fromUnix
	if fromUnix <= cursor.RegisteredUnix {
		cursor.BackfillDone = true
	}
	if err := s.Ledger.SaveSyncCursor(*cursor); err != nil {
		return n, cursor.BackfillDone, err
	}
	return n, cursor.BackfillDone, nil
}

func (s *Syncer) backfillTo(cursor store.SyncCursor) (int64, error) {
	if cursor.OldestUnix > 0 {
		return cursor.OldestUnix, nil
	}
	minUnix, _, count, err := s.Ledger.ScrobbleBounds()
	if err != nil {
		return 0, err
	}
	if count > 0 && minUnix > 0 {
		return minUnix, nil
	}
	return s.Now().Unix() + 1, nil
}

func windowFrom(toUnix, registered int64) int64 {
	from := toUnix - int64(backfillWindow/time.Second)
	if registered > 0 && from < registered {
		return registered
	}
	if from < 0 {
		return 0
	}
	return from
}
