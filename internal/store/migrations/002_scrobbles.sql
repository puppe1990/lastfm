-- up
CREATE TABLE scrobbles (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  played_unix INTEGER NOT NULL,
  track TEXT NOT NULL,
  artist TEXT NOT NULL,
  album TEXT NOT NULL DEFAULT '',
  url TEXT NOT NULL DEFAULT '',
  image TEXT NOT NULL DEFAULT '',
  UNIQUE (played_unix, artist, track, album)
);

CREATE INDEX scrobbles_played_unix ON scrobbles (played_unix DESC);
CREATE INDEX scrobbles_artist ON scrobbles (artist);

CREATE TABLE scrobble_sync (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  newest_unix INTEGER NOT NULL DEFAULT 0,
  oldest_unix INTEGER NOT NULL DEFAULT 0,
  registered_unix INTEGER NOT NULL DEFAULT 0,
  backfill_done INTEGER NOT NULL DEFAULT 0,
  last_error TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT ''
);

-- down
DROP TABLE IF EXISTS scrobble_sync;
DROP TABLE IF EXISTS scrobbles;
