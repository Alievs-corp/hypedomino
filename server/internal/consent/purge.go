package consent

import (
	"context"
	"log/slog"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/account"
	"github.com/Alievs-corp/hypedomino/server/internal/metrics"
)

const purgeBatch = 500

func RetentionCutoff(now time.Time, years int) time.Time {
	return time.Date(now.UTC().Year()-years, time.January, 1, 0, 0, 0, 0, time.UTC)
}

func (s *Store) PurgeDeletedBefore(ctx context.Context, cutoff time.Time) (int64, error) {
	var purged int64
	for {
		tag, err := s.pool.Exec(ctx, `delete from consent_records where id in (
			select c.id from consent_records c join accounts a on a.id = c.account_id
			where a.state = $1 and a.deleted_at < $2 limit $3)`, account.StateDeleted, cutoff.UTC(), purgeBatch)
		if err != nil {
			return purged, err
		}
		purged += tag.RowsAffected()
		if tag.RowsAffected() < purgeBatch {
			return purged, nil
		}
	}
}

type Retention func() (years int, known bool)

type Purge struct {
	store     *Store
	retention Retention
	logger    *slog.Logger
	runs      *metrics.Counter
	purged    *metrics.Counter
}

func NewPurge(store *Store, retention Retention, logger *slog.Logger) *Purge {
	return &Purge{
		store:     store,
		retention: retention,
		logger:    logger,
		runs:      metrics.NewCounter("hd_consent_purge_runs_total", "Consent record purge runs by result: ok, failed, or waiting while the privacy config is not loaded.", metrics.Labels{"result"}),
		purged:    metrics.NewCounter("hd_consent_records_purged_total", "Consent records of deleted accounts deleted after ConsentRecordRetentionYears.", metrics.Labels{}),
	}
}

func (p *Purge) Runs() *metrics.Counter {
	return p.runs
}

func (p *Purge) Purged() *metrics.Counter {
	return p.purged
}

func (p *Purge) Once(ctx context.Context, now time.Time) (int64, error) {
	years, known := p.retention()
	if !known {
		p.runs.Inc(metrics.Labels{"waiting"})
		return 0, nil
	}
	purged, err := p.store.PurgeDeletedBefore(ctx, RetentionCutoff(now, years))
	p.purged.Add(metrics.Labels{}, uint64(purged))
	if err != nil {
		p.runs.Inc(metrics.Labels{"failed"})
		return purged, err
	}
	p.runs.Inc(metrics.Labels{"ok"})
	return purged, nil
}

func (p *Purge) Run(ctx context.Context, every time.Duration, now func() time.Time) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		purged, err := p.Once(ctx, now())
		if err != nil {
			p.logger.Error("consent record purge failed", "error", err)
		}
		if purged > 0 {
			p.logger.Info("consent records deleted", "count", purged)
		}
	}
}
