package adminapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/rulesconfig"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

const RulesHistoryLimit = 50

type ruleKeyView struct {
	Key     string `json:"key"`
	Type    string `json:"type"`
	Unit    string `json:"unit"`
	Min     int    `json:"min"`
	Max     int    `json:"max"`
	Default int    `json:"default"`
	Value   int    `json:"value"`
	Set     bool   `json:"set"`
}

type rulesView struct {
	ConfigVersion int64         `json:"config_version"`
	Keys          []ruleKeyView `json:"keys"`
}

type rulesVersionView struct {
	ConfigVersion int64              `json:"config_version"`
	PublishedAt   time.Time          `json:"published_at"`
	Values        rulesconfig.Tuning `json:"values"`
}

type rulesHistoryView struct {
	Versions []rulesVersionView `json:"versions"`
}

type publishRulesRequest struct {
	Values map[string]json.RawMessage `json:"values"`
	changeNote
}

type rollbackRulesRequest struct {
	Version int64 `json:"version"`
	changeNote
}

func viewRules(version int64, tuning rulesconfig.Tuning) rulesView {
	keys := make([]ruleKeyView, 0, len(rulesconfig.Keys()))
	for _, key := range rulesconfig.Keys() {
		value, set := tuning[key.Name]
		if !set {
			value = key.Default()
		}
		keys = append(keys, ruleKeyView{Key: key.Name, Type: "integer", Unit: key.Unit, Min: key.Min, Max: key.Max, Default: key.Default(), Value: value, Set: set})
	}
	return rulesView{ConfigVersion: version, Keys: keys}
}

func latestRules(ctx context.Context, db snapshot.Querier) (int64, []byte, rulesconfig.Tuning, error) {
	version, raw, err := snapshot.Latest(ctx, db, rulesconfig.Domain)
	if errors.Is(err, snapshot.ErrNone) {
		return 0, nil, rulesconfig.Tuning{}, nil
	}
	if err != nil {
		return 0, nil, nil, err
	}
	tuning, err := rulesconfig.Parse(raw)
	return version, raw, tuning, err
}

func (h *Handler) rules(w http.ResponseWriter, r *http.Request, p principal) {
	version, _, tuning, err := latestRules(r.Context(), h.pool)
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, viewRules(version, tuning))
}

func (h *Handler) rulesHistory(w http.ResponseWriter, r *http.Request, p principal) {
	history, err := snapshot.History(r.Context(), h.pool, rulesconfig.Domain, RulesHistoryLimit)
	if err != nil {
		h.refuse(w, err)
		return
	}
	versions := make([]rulesVersionView, 0, len(history))
	for _, published := range history {
		var values rulesconfig.Tuning
		if err := json.Unmarshal(published.Content, &values); err != nil {
			h.refuse(w, err)
			return
		}
		versions = append(versions, rulesVersionView{ConfigVersion: published.Version, PublishedAt: published.At, Values: values})
	}
	reply(w, http.StatusOK, rulesHistoryView{Versions: versions})
}

func (h *Handler) publishRules(w http.ResponseWriter, r *http.Request, p principal) {
	var request publishRulesRequest
	if !decode(w, r, &request) {
		return
	}
	if request.Values == nil {
		request.Values = map[string]json.RawMessage{}
	}
	h.replaceRules(w, r, p, "publish", marshal(request.Values), request.changeNote, nil)
}

func (h *Handler) rollbackRules(w http.ResponseWriter, r *http.Request, p principal) {
	var request rollbackRulesRequest
	if !decode(w, r, &request) {
		return
	}
	target, err := snapshot.Version(r.Context(), h.pool, rulesconfig.Domain, request.Version)
	if errors.Is(err, snapshot.ErrNone) {
		err = errNoSuchVersion
	}
	if err != nil {
		h.refuse(w, err)
		return
	}
	h.replaceRules(w, r, p, "rollback", target.Content, request.changeNote, h.mayRollBack)
}

type sensitiveCheck func(p principal, note changeNote, now time.Time) error

func (h *Handler) replaceRules(w http.ResponseWriter, r *http.Request, p principal, action string, raw []byte, note changeNote, check sensitiveCheck) {
	tuning, err := rulesconfig.Parse(raw)
	if err != nil {
		h.refuse(w, err)
		return
	}
	var version int64
	err = h.audited(r.Context(), p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
		if check != nil {
			if err := check(p, note, now); err != nil {
				return nil, err
			}
		}
		if err := lockDomain(r.Context(), tx, rulesconfig.Domain); err != nil {
			return nil, err
		}
		_, before, _, err := latestRules(r.Context(), tx)
		if err != nil {
			return nil, err
		}
		if version, err = snapshot.Publish(r.Context(), tx, rulesconfig.Domain, raw); err != nil {
			return nil, err
		}
		entry := audit.Entry{
			Action: action, Object: audit.Object{Type: rbac.GameConfig, ID: rulesconfig.Domain}, After: json.RawMessage(raw),
			ConfigVersion: version, Reason: note.Reason, Ref: note.Ref,
		}
		if before != nil {
			entry.Before = before
		}
		return []audit.Entry{entry}, nil
	})
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, viewRules(version, tuning))
}
