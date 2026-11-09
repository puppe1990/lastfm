package jobs

import (
	"context"
	"log"

	caisjobs "github.com/puppe1990/amarra-cais/pkg/cais/jobs"

	"github.com/puppe1990/lastfm/internal/scrobbles"
	"github.com/puppe1990/lastfm/internal/store"
)

// KindSyncScrobbles is the Amarra job that copies Last.fm history into SQLite.
const KindSyncScrobbles = "SyncScrobbles"

// PerformSyncScrobbles pulls the last day, then backfills older pages until registered.
func PerformSyncScrobbles(s store.Store, src scrobbles.Source) caisjobs.Handler {
	return func(ctx context.Context, _ []byte) error {
		syncer := scrobbles.Syncer{
			Source:         src,
			Ledger:         s,
			BackfillBudget: scrobbles.DefaultBackfillBudget,
		}
		rep, err := syncer.Run(ctx)
		if err != nil {
			return err
		}
		log.Printf("scrobbles sync inserted=%d stored=%d backfill_done=%v", rep.Inserted, rep.Stored, rep.BackfillDone)
		return nil
	}
}
