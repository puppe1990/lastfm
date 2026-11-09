package db

import (
	"testing"

	caisjobs "github.com/puppe1990/amarra-cais/pkg/cais/jobs"

	appjobs "github.com/puppe1990/lastfm/internal/jobs"
	"github.com/puppe1990/lastfm/internal/store"
)

func TestRunSeeds_RegistersFiveMinuteSync(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	if err := RunSeeds(s); err != nil {
		t.Fatal(err)
	}
	tasks, err := caisjobs.NewStore(s.DB()).ListRecurring(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 1 {
		t.Fatalf("recurring = %d, want 1", len(tasks))
	}
	if tasks[0].Kind != appjobs.KindSyncScrobbles || tasks[0].Cron != "*/5 * * * *" {
		t.Fatalf("task = %+v", tasks[0])
	}
}
