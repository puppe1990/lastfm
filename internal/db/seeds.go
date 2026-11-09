package db

import (
	"context"

	caisjobs "github.com/puppe1990/amarra-cais/pkg/cais/jobs"

	appjobs "github.com/puppe1990/lastfm/internal/jobs"
	"github.com/puppe1990/lastfm/internal/store"
)

// RunSeeds registers the recurring scrobble sync. Safe to run multiple times.
func RunSeeds(s store.Store) error {
	jobStore := caisjobs.NewStore(s.DB())
	if err := caisjobs.EnsureSchema(s.DB()); err != nil {
		return err
	}
	if err := jobStore.UpsertRecurring(context.Background(), caisjobs.RecurringOptions{
		Kind: appjobs.KindSyncScrobbles,
		Cron: "*/5 * * * *",
	}); err != nil {
		return err
	}
	// cais:recurring-seeds
	// cais:seeds
	return nil
}
