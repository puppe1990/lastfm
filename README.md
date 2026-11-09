# Ouvindo

Personal Last.fm listening ledger for [puppe90](https://www.last.fm/user/puppe90). Go + [Amarra](https://github.com/puppe1990/amarra-cais). The API key stays on the server.

Live: https://lastfm.apps.gestaobem.com

## Pages

| Path | Last.fm method |
| --- | --- |
| `/` | `user.getInfo` + `user.getRecentTracks` |
| `/artistas` | `user.getTopArtists` |
| `/albuns` | `user.getTopAlbums` |
| `/faixas` | `user.getTopTracks` |
| `/amadas` | `user.getLovedTracks` |

Chart windows: `7day`, `1month`, `3month`, `6month`, `12month`, `overall` (`?period=`). Responses are cached in memory for 45 seconds.

Home charts read SQLite once the ledger covers 7 days. Job `SyncScrobbles` runs on boot and every 5 minutes: last day first, then backfill from account creation. In-process worker (Cleat only runs `bin/server`).

## Run

```bash
cp .env.example .env   # set LASTFM_API_KEY
amarra-cais install
amarra-cais dev        # http://localhost:8080
```

Tests (from this directory, real Go — RTK intercepts a bare `go test`):

```bash
/opt/homebrew/Cellar/go/1.26.4/bin/go test ./... -count=1
make ci
```

## Env

| Key | Notes |
| --- | --- |
| `LASTFM_API_KEY` | required |
| `LASTFM_USER` | defaults to `puppe90` |
| `LOCALE` | `pt` in production |
| `CSP_IMG_SRC` | Last.fm Fastly hosts |
| `ADMIN_TOKEN` | required when `ENV=production` |

Deploy: `cleat__deploy app=lastfm ref=main`. noindex.
