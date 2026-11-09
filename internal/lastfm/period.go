package lastfm

// Periods is the Last.fm chart window set, in the order the UI shows.
var Periods = []string{"7day", "1month", "3month", "6month", "12month", "overall"}

// NormalizePeriod keeps a known Last.fm period and otherwise uses 7day.
func NormalizePeriod(raw string) string {
	for _, p := range Periods {
		if raw == p {
			return p
		}
	}
	return "7day"
}
