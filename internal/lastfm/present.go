package lastfm

import (
	"strconv"
	"time"
)

// FormatPlaycount is a pt-BR grouped integer (82197 → 82.197).
func FormatPlaycount(n int) string {
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	s := strconv.Itoa(n)
	out := make([]byte, 0, len(s)+len(s)/3)
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, s[i])
	}
	return sign + string(out)
}

// FormatWhen is a short BRT clock for a unix scrobble. Empty if t is zero.
func FormatWhen(unix int64, now time.Time) string {
	if unix <= 0 {
		return ""
	}
	loc := now.Location()
	if loc == nil || loc == time.UTC {
		if br, err := time.LoadLocation("America/Sao_Paulo"); err == nil {
			loc = br
		}
	}
	return time.Unix(unix, 0).In(loc).Format("02/01 15:04")
}

// FormatRegisteredYear is the Last.fm join year, or empty.
func FormatRegisteredYear(unix int64) string {
	if unix <= 0 {
		return ""
	}
	return strconv.Itoa(time.Unix(unix, 0).UTC().Year())
}
