package lastfm

import "time"

var weekdayPT = [...]string{"dom", "seg", "ter", "qua", "qui", "sex", "sáb"}

// CountBar is one column of a Last.fm-style histogram (SVG viewBox height 100).
type CountBar struct {
	Label      string
	Count      int
	CountLabel string
	Pct        int
	Peak       bool
	X, Y, W, H int
}

// ActivityDays buckets scrobbles into `days` BRT calendar days ending today.
func ActivityDays(tracks []Track, now time.Time, days int) []CountBar {
	if days < 1 {
		days = 7
	}
	loc := chartClock(now)
	now = now.In(loc)
	start := midnight(now.AddDate(0, 0, -(days - 1)))
	counts := make([]int, days)
	for _, tr := range tracks {
		if skipScrobble(tr) {
			continue
		}
		t := time.Unix(tr.PlayedUnix, 0).In(loc)
		if t.Before(start) {
			continue
		}
		i := int(midnight(t).Sub(start).Hours() / 24)
		if i >= 0 && i < days {
			counts[i]++
		}
	}
	bars := make([]CountBar, days)
	for i, n := range counts {
		day := start.AddDate(0, 0, i)
		bars[i] = CountBar{
			Label:      weekdayPT[day.Weekday()],
			Count:      n,
			CountLabel: FormatPlaycount(n),
		}
	}
	layoutBars(bars, 2, 8)
	return bars
}

// ClockHours buckets the last 7 BRT days into 24 hour-of-day columns.
func ClockHours(tracks []Track, now time.Time) []CountBar {
	loc := chartClock(now)
	now = now.In(loc)
	start := midnight(now.AddDate(0, 0, -6))
	counts := make([]int, 24)
	for _, tr := range tracks {
		if skipScrobble(tr) {
			continue
		}
		t := time.Unix(tr.PlayedUnix, 0).In(loc)
		if t.Before(start) {
			continue
		}
		counts[t.Hour()]++
	}
	bars := make([]CountBar, 24)
	for h, n := range counts {
		bars[h] = CountBar{
			Label:      FormatPlaycount(h),
			Count:      n,
			CountLabel: FormatPlaycount(n),
		}
	}
	layoutBars(bars, 2, 8)
	return bars
}

// RelativePercents scales counts to 0–100 of the max. A non-zero bar is at least 8.
func RelativePercents(counts []int) []int {
	if counts == nil {
		return nil
	}
	max := 0
	for _, n := range counts {
		if n > max {
			max = n
		}
	}
	out := make([]int, len(counts))
	if max == 0 {
		return out
	}
	for i, n := range counts {
		if n <= 0 {
			continue
		}
		pct := n * 100 / max
		if pct < 8 {
			pct = 8
		}
		out[i] = pct
	}
	return out
}

func layoutBars(bars []CountBar, gap, width int) {
	pcts := make([]int, len(bars))
	for i := range bars {
		pcts[i] = bars[i].Count
	}
	pcts = RelativePercents(pcts)
	peak := 0
	for _, n := range bars {
		if n.Count > peak {
			peak = n.Count
		}
	}
	for i := range bars {
		bars[i].Pct = pcts[i]
		bars[i].Peak = peak > 0 && bars[i].Count == peak
		bars[i].W = width
		bars[i].H = pcts[i]
		bars[i].Y = 100 - pcts[i]
		bars[i].X = i * (width + gap)
	}
}

func skipScrobble(tr Track) bool {
	return tr.NowPlaying || tr.PlayedUnix <= 0
}

func midnight(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func chartClock(now time.Time) *time.Location {
	if now.Location() != nil && now.Location() != time.UTC {
		return now.Location()
	}
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("BRT", -3*3600)
	}
	return loc
}
