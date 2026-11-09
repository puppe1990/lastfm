package lastfm

import (
	"testing"
	"time"
)

func TestFormatPlaycount_GroupsThousandsWithDot(t *testing.T) {
	if got := FormatPlaycount(82197); got != "82.197" {
		t.Errorf("82197 = %q", got)
	}
	if got := FormatPlaycount(12); got != "12" {
		t.Errorf("12 = %q", got)
	}
	if got := FormatPlaycount(0); got != "0" {
		t.Errorf("0 = %q", got)
	}
}

func TestFormatWhen_UsesBRTClock(t *testing.T) {
	loc := time.FixedZone("BRT", -3*3600)
	now := time.Date(2026, 11, 9, 18, 0, 0, 0, loc)
	played := time.Date(2026, 11, 9, 14, 48, 0, 0, loc)
	if got := FormatWhen(played.Unix(), now); got != "09/11 14:48" {
		t.Errorf("when = %q", got)
	}
	if FormatWhen(0, now) != "" {
		t.Error("zero unix should be empty")
	}
}

func TestFormatRegisteredYear(t *testing.T) {
	if got := FormatRegisteredYear(1201408060); got != "2008" {
		t.Errorf("year = %q", got)
	}
	if got := FormatRegisteredYear(0); got != "" {
		t.Errorf("zero = %q", got)
	}
}
