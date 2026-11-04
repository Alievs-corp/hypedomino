package adminapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/rulesconfig"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

var rulesDomain = configDomain{name: rulesconfig.Domain, object: rbac.GameConfig, accept: func(_, after []byte) error {
	_, err := rulesconfig.Parse(after)
	return err
}}

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

type rollbackRequest struct {
	Version int64 `json:"version"`
	changeNote
}

func viewRules(version int64, raw []byte) (rulesView, error) {
	tuning := rulesconfig.Tuning{}
	if raw != nil {
		var err error
		if tuning, err = rulesconfig.Parse(raw); err != nil {
			return rulesView{}, err
		}
	}
	keys := make([]ruleKeyView, 0, len(rulesconfig.Keys()))
	for _, key := range rulesconfig.Keys() {
		value, set := tuning[key.Name]
		if !set {
			value = key.Default()
		}
		keys = append(keys, ruleKeyView{Key: key.Name, Type: "integer", Unit: key.Unit, Min: key.Min, Max: key.Max, Default: key.Default(), Value: value, Set: set})
	}
	return rulesView{ConfigVersion: version, Keys: keys}, nil
}

func (h *Handler) replyRules(w http.ResponseWriter, version int64, raw []byte) {
	view, err := viewRules(version, raw)
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, view)
}

func (h *Handler) rules(w http.ResponseWriter, r *http.Request, p principal) {
	version, raw, err := latestSnapshot(r.Context(), h.pool, rulesconfig.Domain)
	if err != nil {
		h.refuse(w, err)
		return
	}
	h.replyRules(w, version, raw)
}

func (h *Handler) rulesHistory(w http.ResponseWriter, r *http.Request, p principal) {
	history, err := snapshot.History(r.Context(), h.pool, rulesconfig.Domain, ConfigHistoryLimit)
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
	raw := marshal(request.Values)
	version, _, err := h.replaceSnapshot(r.Context(), p, rulesDomain, "publish", raw, request.changeNote, nil)
	if err != nil {
		h.refuse(w, err)
		return
	}
	h.replyRules(w, version, raw)
}

func (h *Handler) rollbackRules(w http.ResponseWriter, r *http.Request, p principal) {
	var request rollbackRequest
	if !decode(w, r, &request) {
		return
	}
	raw, err := h.earlierSnapshot(r.Context(), rulesconfig.Domain, request.Version)
	if err != nil {
		h.refuse(w, err)
		return
	}
	version, _, err := h.replaceSnapshot(r.Context(), p, rulesDomain, "rollback", raw, request.changeNote, h.mayRollBack)
	if err != nil {
		h.refuse(w, err)
		return
	}
	h.replyRules(w, version, raw)
}
