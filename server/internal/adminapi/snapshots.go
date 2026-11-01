package adminapi

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

const ConfigHistoryLimit = 50

type configDomain struct {
	name   string
	object rbac.Object
	accept func(before, after []byte) error
}

type sensitiveCheck func(p principal, note changeNote, now time.Time) error

func latestSnapshot(ctx context.Context, db snapshot.Querier, domain string) (int64, []byte, error) {
	version, raw, err := snapshot.Latest(ctx, db, domain)
	if errors.Is(err, snapshot.ErrNone) {
		return 0, nil, nil
	}
	return version, raw, err
}

func (h *Handler) earlierSnapshot(ctx context.Context, domain string, version int64) ([]byte, error) {
	target, err := snapshot.Version(ctx, h.pool, domain, version)
	if errors.Is(err, snapshot.ErrNone) {
		return nil, errNoSuchVersion
	}
	if err != nil {
		return nil, err
	}
	return target.Content, nil
}

func (h *Handler) replaceSnapshot(ctx context.Context, p principal, d configDomain, action string, raw []byte, note changeNote, check sensitiveCheck) (int64, error) {
	var version int64
	err := h.audited(ctx, p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
		if check != nil {
			if err := check(p, note, now); err != nil {
				return nil, err
			}
		}
		if err := lockDomain(ctx, tx, d.name); err != nil {
			return nil, err
		}
		_, before, err := latestSnapshot(ctx, tx, d.name)
		if err != nil {
			return nil, err
		}
		if err := d.accept(before, raw); err != nil {
			return nil, err
		}
		if version, err = snapshot.Publish(ctx, tx, d.name, raw); err != nil {
			return nil, err
		}
		entry := audit.Entry{
			Action: action, Object: audit.Object{Type: d.object, ID: d.name}, After: json.RawMessage(raw),
			Before: before, ConfigVersion: version, Reason: note.Reason, Ref: note.Ref,
		}
		return []audit.Entry{entry}, nil
	})
	return version, err
}
