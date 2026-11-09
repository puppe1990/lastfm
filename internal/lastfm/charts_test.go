package lastfm

import (
	"testing"
	"time"
)

func brtDate(y int, m time.Month, d, h int) time.Time {
	loc := time.FixedZone("BRT", -3*3600)
	return time.Date(y, m, d, h, 0, 0, 0, loc)
}

func TestActivityDays_FillsSevenBRTDaysAndSkipsNowPlaying(t *testing.T) {
	now := brtDate(2026, 11, 9, 18)
	tracks := []Track{
		{PlayedUnix: brtDate(2026, 11, 9, 10).Unix()},
		{PlayedUnix: brtDate(2026, 11, 9, 11).Unix()},
		{PlayedUnix: brtDate(2026, 11, 9, 11).Unix()},
		{PlayedUnix: brtDate(2026, 11, 8, 22).Unix()},
		{PlayedUnix: brtDate(2026, 11, 2, 8).Unix()}, // before the 7-day window
		{NowPlaying: true, PlayedUnix: now.Unix()},
		{PlayedUnix: 0},
	}

	bars := ActivityDays(tracks, now, 7)
	if len(bars) != 7 {
		t.Fatalf("len = %d, want 7", len(bars))
	}
	if bars[0].Label != "ter" || bars[6].Label != "seg" {
		t.Fatalf("labels = %q…%q", bars[0].Label, bars[6].Label)
	}
	if bars[5].Count != 1 {
		t.Errorf("sunday count = %d, want 1", bars[5].Count)
	}
	if bars[6].Count != 3 {
		t.Errorf("monday count = %d, want 3", bars[6].Count)
	}
	if !bars[6].Peak {
		t.Error("today should be the peak")
	}
	if bars[6].Pct != 100 {
		t.Errorf("peak pct = %d", bars[6].Pct)
	}
	if bars[6].H != 100 || bars[6].Y != 0 {
		t.Errorf("peak svg H=%d Y=%d", bars[6].H, bars[6].Y)
	}
	if bars[0].Count != 0 || bars[0].H != 0 {
		t.Errorf("empty day count=%d H=%d", bars[0].Count, bars[0].H)
	}
}

func TestClockHours_BucketsBRTHours(t *testing.T) {
	now := brtDate(2026, 11, 9, 18)
	tracks := []Track{
		{PlayedUnix: brtDate(2026, 11, 9, 0).Unix()},
		{PlayedUnix: brtDate(2026, 11, 8, 18).Unix()},
		{PlayedUnix: brtDate(2026, 11, 8, 18).Unix()},
		{PlayedUnix: brtDate(2026, 11, 1, 18).Unix()}, // too old
		{NowPlaying: true, PlayedUnix: brtDate(2026, 11, 9, 18).Unix()},
	}

	bars := ClockHours(tracks, now)
	if len(bars) != 24 {
		t.Fatalf("len = %d, want 24", len(bars))
	}
	if bars[0].Count != 1 {
		t.Errorf("hour 0 = %d", bars[0].Count)
	}
	if bars[18].Count != 2 || !bars[18].Peak {
		t.Errorf("hour 18 count=%d peak=%v", bars[18].Count, bars[18].Peak)
	}
	if bars[18].Label != "18" {
		t.Errorf("label = %q", bars[18].Label)
	}
}

func TestRelativePercents_ScalesToMaxWithVisibleStub(t *testing.T) {
	got := RelativePercents([]int{100, 50, 0, 1})
	want := []int{100, 50, 0, 8}
	if len(got) != len(want) {
		t.Fatalf("len = %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("pct[%d] = %d, want %d", i, got[i], want[i])
		}
	}
	if RelativePercents(nil) != nil {
		t.Error("nil in should be nil out")
	}
}
