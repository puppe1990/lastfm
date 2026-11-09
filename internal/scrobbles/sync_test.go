package scrobbles

import (
	"context"
	"testing"
	"time"

	"github.com/puppe1990/lastfm/internal/lastfm"
	"github.com/puppe1990/lastfm/internal/store"
)

type memSource struct {
	user lastfm.User
	all  []lastfm.Track
}

func (m *memSource) User(context.Context) (lastfm.User, error) {
	return m.user, nil
}

func (m *memSource) RecentTracksRange(_ context.Context, fromUnix, toUnix int64, page, limit int) (lastfm.RecentPage, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 200
	}
	var hit []lastfm.Track
	for _, tr := range m.all {
		if fromUnix > 0 && tr.PlayedUnix <= fromUnix {
			continue
		}
		if toUnix > 0 && tr.PlayedUnix >= toUnix {
			continue
		}
		hit = append(hit, tr)
	}
	total := len(hit)
	pages := (total + limit - 1) / limit
	if pages < 1 {
		pages = 1
	}
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return lastfm.RecentPage{Tracks: hit[start:end], Page: page, TotalPages: pages, Total: total}, nil
}

func testNow() time.Time {
	return time.Date(2026, 11, 9, 21, 0, 0, 0, time.UTC)
}

func TestSyncer_PullsLastDayThenBackfillsOlder(t *testing.T) {
	now := testNow()
	registered := now.AddDate(-10, 0, 0).Unix()
	fresh := lastfm.Track{Name: "Bleed", Artist: "Meshuggah", PlayedUnix: now.Add(-2 * time.Hour).Unix()}
	week := lastfm.Track{Name: "In Waves", Artist: "Trivium", PlayedUnix: now.Add(-48 * time.Hour).Unix()}
	old := lastfm.Track{Name: "Ghost of Perdition", Artist: "Opeth", PlayedUnix: now.AddDate(-2, 0, 0).Unix()}
	playing := lastfm.Track{Name: "now", Artist: "NIN", NowPlaying: true, PlayedUnix: now.Unix()}

	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	syncer := Syncer{
		Source: &memSource{
			user: lastfm.User{Name: "puppe90", RegisteredUnix: registered, Playcount: 3},
			all:  []lastfm.Track{playing, fresh, week, old},
		},
		Ledger:         s,
		Now:            func() time.Time { return now },
		Sleep:          func(time.Duration) {},
		PageSize:       200,
		BackfillBudget: time.Hour,
	}
	rep, err := syncer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rep.Inserted != 3 {
		t.Fatalf("inserted = %d, want 3", rep.Inserted)
	}
	if !rep.BackfillDone {
		t.Fatal("backfill should finish")
	}
	_, _, count, err := s.ScrobbleBounds()
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("stored = %d, want 3", count)
	}
	got, err := s.ListScrobblesBetween(0, now.Unix()+1)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, tr := range got {
		names[tr.Name] = true
	}
	for _, want := range []string{"Bleed", "In Waves", "Ghost of Perdition"} {
		if !names[want] {
			t.Errorf("missing %q", want)
		}
	}
}

func TestSyncer_SkipsBackfillWhenAlreadyDone(t *testing.T) {
	now := testNow()
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := s.SaveSyncCursor(store.SyncCursor{BackfillDone: true, RegisteredUnix: 1}); err != nil {
		t.Fatal(err)
	}

	fresh := lastfm.Track{Name: "today", Artist: "A", PlayedUnix: now.Add(-time.Hour).Unix()}
	old := lastfm.Track{Name: "ancient", Artist: "B", PlayedUnix: now.AddDate(-3, 0, 0).Unix()}
	syncer := Syncer{
		Source: &memSource{
			user: lastfm.User{RegisteredUnix: 1},
			all:  []lastfm.Track{fresh, old},
		},
		Ledger:         s,
		Now:            func() time.Time { return now },
		Sleep:          func(time.Duration) {},
		PageSize:       200,
		BackfillBudget: time.Hour,
	}
	rep, err := syncer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rep.Inserted != 1 {
		t.Fatalf("inserted = %d, want 1 last-day only", rep.Inserted)
	}
	got, err := s.ListScrobblesBetween(0, now.Unix()+1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "today" {
		t.Fatalf("stored = %+v", got)
	}
}

type spySource struct {
	inner *memSource
	calls [][2]int64
}

func (s *spySource) User(ctx context.Context) (lastfm.User, error) {
	return s.inner.User(ctx)
}

func (s *spySource) RecentTracksRange(ctx context.Context, fromUnix, toUnix int64, page, limit int) (lastfm.RecentPage, error) {
	if page == 1 {
		s.calls = append(s.calls, [2]int64{fromUnix, toUnix})
	}
	return s.inner.RecentTracksRange(ctx, fromUnix, toUnix, page, limit)
}

func TestSyncer_BackfillUsesThirtyDayWindows(t *testing.T) {
	now := testNow()
	registered := now.AddDate(-2, 0, 0).Unix()
	src := &spySource{inner: &memSource{
		user: lastfm.User{Name: "puppe90", RegisteredUnix: registered},
		all: []lastfm.Track{
			{Name: "today", Artist: "A", PlayedUnix: now.Add(-2 * time.Hour).Unix()},
			{Name: "month", Artist: "B", PlayedUnix: now.Add(-20 * 24 * time.Hour).Unix()},
			{Name: "older", Artist: "C", PlayedUnix: now.Add(-50 * 24 * time.Hour).Unix()},
		},
	}}
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	syncer := Syncer{
		Source:         src,
		Ledger:         s,
		Now:            func() time.Time { return now },
		Sleep:          func(time.Duration) {},
		PageSize:       200,
		BackfillBudget: time.Hour,
	}
	rep, err := syncer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rep.Inserted != 3 {
		t.Fatalf("inserted = %d, want 3", rep.Inserted)
	}
	if !rep.BackfillDone {
		t.Fatal("backfill should finish")
	}
	if len(src.calls) < 2 {
		t.Fatalf("calls = %d, want last-day plus backfill windows", len(src.calls))
	}
	if src.calls[0][1] != 0 {
		t.Fatalf("last-day to = %d, want 0 (omit to)", src.calls[0][1])
	}
	const maxSpan = 31 * 24 * 3600
	for i, c := range src.calls[1:] {
		if c[1] <= 0 {
			continue
		}
		span := c[1] - c[0]
		if span > maxSpan {
			t.Errorf("window %d span %d s (from=%d to=%d) exceeds 31d", i, span, c[0], c[1])
		}
	}
}

func TestSyncer_ResumesBackfillAfterBudget(t *testing.T) {
	now := testNow()
	registered := now.AddDate(-5, 0, 0).Unix()
	offsets := []time.Duration{
		time.Hour,
		2 * time.Hour,
		48 * time.Hour,
		72 * time.Hour,
		96 * time.Hour,
	}
	var all []lastfm.Track
	for i, off := range offsets {
		all = append(all, lastfm.Track{
			Name:       "t",
			Artist:     "A",
			Album:      lastfm.FormatPlaycount(i),
			PlayedUnix: now.Add(-off).Unix(),
		})
	}

	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	syncer := Syncer{
		Source: &memSource{
			user: lastfm.User{RegisteredUnix: registered},
			all:  all,
		},
		Ledger:         s,
		Now:            func() time.Time { return now },
		Sleep:          func(time.Duration) {},
		PageSize:       2,
		BackfillBudget: 0,
	}
	rep, err := syncer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rep.BackfillDone {
		t.Fatal("budget 0 should leave backfill open")
	}
	if _, _, count, _ := s.ScrobbleBounds(); count != 2 {
		t.Fatalf("first run stored %d, want last-day page of 2", count)
	}

	syncer.BackfillBudget = time.Hour
	rep, err = syncer.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !rep.BackfillDone {
		t.Fatal("second run should finish backfill")
	}
	if _, _, count, _ := s.ScrobbleBounds(); count != 5 {
		t.Fatalf("stored = %d, want 5", count)
	}
}
