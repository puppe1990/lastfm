package jobs

import (
	"database/sql"

	caisjobs "github.com/puppe1990/amarra-cais/pkg/cais/jobs"

	"github.com/puppe1990/lastfm/internal/scrobbles"
	"github.com/puppe1990/lastfm/internal/store"
)

// RegisterAll wires app and framework jobs into the worker registry.
func RegisterAll(reg *caisjobs.Registry, db *sql.DB, s store.Store, src scrobbles.Source) {
	reg.Register(caisjobs.KindPruneSessions, caisjobs.PruneSessionsHandler(db))
	reg.Register(caisjobs.KindPruneFinished, caisjobs.PruneFinishedHandler(db))
	if src != nil {
		reg.Register(KindSyncScrobbles, PerformSyncScrobbles(s, src))
	}
	// cais:jobs-register
}
