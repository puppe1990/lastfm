package app

import (
	"context"
	"fmt"
	"log"

	caisjobs "github.com/puppe1990/amarra-cais/pkg/cais/jobs"

	appdb "github.com/puppe1990/lastfm/internal/db"
	appjobs "github.com/puppe1990/lastfm/internal/jobs"
	"github.com/puppe1990/lastfm/internal/scrobbles"
	"github.com/puppe1990/lastfm/internal/store"
)

func bootWorker(cfgDBPath, env string, s store.Store, src scrobbles.Source) (*caisjobs.Worker, store.Store, error) {
	if src == nil {
		return nil, nil, nil
	}
	if err := caisjobs.EnsureSchema(s.DB()); err != nil {
		return nil, nil, fmt.Errorf("jobs schema: %w", err)
	}
	if err := appdb.RunSeeds(s); err != nil {
		return nil, nil, fmt.Errorf("job seeds: %w", err)
	}

	reg := caisjobs.NewRegistry()
	appjobs.RegisterAll(reg, s.DB(), s, src)

	var heartbeat store.Store
	hbStore := caisjobs.NewStore(s.DB())
	if cfgDBPath != "" && cfgDBPath != ":memory:" {
		hb, err := store.NewSQLiteStore(cfgDBPath, env)
		if err != nil {
			return nil, nil, fmt.Errorf("jobs heartbeat store: %w", err)
		}
		heartbeat = hb
		hbStore = caisjobs.NewStore(hb.DB())
	}

	worker := caisjobs.NewWorker(caisjobs.WorkerConfig{
		Store:          caisjobs.NewStore(s.DB()),
		HeartbeatStore: hbStore,
		Registry:       reg,
		Concurrency:    1,
	})

	if _, err := caisjobs.Enqueue(context.Background(), caisjobs.NewStore(s.DB()), caisjobs.Options{
		Kind: appjobs.KindSyncScrobbles,
	}); err != nil {
		if heartbeat != nil {
			_ = heartbeat.Close()
		}
		return nil, nil, fmt.Errorf("enqueue scrobble sync: %w", err)
	}
	log.Printf("jobs: queued %s and cron */5", appjobs.KindSyncScrobbles)
	return worker, heartbeat, nil
}
