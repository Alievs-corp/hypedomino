package adminapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Alievs-corp/hypedomino/server/internal/audit"
	"github.com/Alievs-corp/hypedomino/server/internal/contentgate"
	"github.com/Alievs-corp/hypedomino/server/internal/cosmetics"
	"github.com/Alievs-corp/hypedomino/server/internal/gamestrings"
	"github.com/Alievs-corp/hypedomino/server/internal/liveops"
	"github.com/Alievs-corp/hypedomino/server/internal/promo"
	"github.com/Alievs-corp/hypedomino/server/internal/rating"
	"github.com/Alievs-corp/hypedomino/server/internal/rbac"
	"github.com/Alievs-corp/hypedomino/server/internal/snapshot"
)

const ContentGateHistoryLimit = 50

var errReportUnknown = errors.New("adminapi: no such content gate report")

type gateVersions struct {
	ConfigVersion   int64                        `json:"config_version"`
	Domains         map[string]int64             `json:"domains"`
	ManifestVersion int                          `json:"manifest_version"`
	StringBundles   map[gamestrings.Language]int `json:"string_bundles"`
}

type gateReport struct {
	ID          string                  `json:"id"`
	Environment string                  `json:"environment"`
	LaunchAt    time.Time               `json:"launch_at"`
	RanAt       time.Time               `json:"ran_at"`
	Runner      string                  `json:"runner"`
	Versions    gateVersions            `json:"versions"`
	Thresholds  []contentgate.Threshold `json:"thresholds"`
	contentgate.Result
}

type gateCheckResult struct {
	ID     string `json:"id"`
	Result string `json:"result"`
}

type gateSummary struct {
	ID          string            `json:"id"`
	Environment string            `json:"environment"`
	LaunchAt    time.Time         `json:"launch_at"`
	RanAt       time.Time         `json:"ran_at"`
	Runner      string            `json:"runner"`
	Verdict     string            `json:"verdict"`
	Stale       bool              `json:"stale"`
	Results     []gateCheckResult `json:"results"`
}

type gateReportView struct {
	Report json.RawMessage `json:"report"`
	Stale  bool            `json:"stale"`
}

func (v gateVersions) staleAgainst(current map[string]int64) bool {
	for _, domain := range contentgate.Domains {
		if v.Domains[domain] != current[domain] {
			return true
		}
	}
	return false
}

func (h *Handler) runContentGate(w http.ResponseWriter, r *http.Request, p principal) {
	var request struct {
		LaunchAt time.Time `json:"launch_at"`
	}
	if !decode(w, r, &request) {
		return
	}
	if request.LaunchAt.IsZero() {
		h.refuse(w, fmt.Errorf("%w: launch_at is required", errMalformed))
		return
	}
	var report gateReport
	err := h.audited(r.Context(), p, func(tx pgx.Tx, now time.Time) ([]audit.Entry, error) {
		var err error
		if report, err = h.gateReport(r.Context(), tx, p, request.LaunchAt.UTC(), now.UTC()); err != nil {
			return nil, err
		}
		if err := storeGateReport(r.Context(), tx, report); err != nil {
			return nil, err
		}
		after := marshal(map[string]any{"environment": report.Environment, "launch_at": report.LaunchAt, "verdict": report.Verdict})
		return []audit.Entry{{Action: string(rbac.Create), Object: audit.Object{Type: rbac.ContentGateReports, ID: report.ID}, After: after, ConfigVersion: report.Versions.ConfigVersion}}, nil
	})
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusCreated, gateReportView{Report: marshal(report)})
}

func (h *Handler) gateReport(ctx context.Context, tx pgx.Tx, p principal, launchAt, now time.Time) (gateReport, error) {
	versions, snapshots, err := gateSnapshots(ctx, tx)
	if err != nil {
		return gateReport{}, err
	}
	thresholds, err := gateThresholds(snapshots[liveops.Domain], versions.Domains[liveops.Domain])
	if err != nil {
		return gateReport{}, err
	}
	bundles, err := h.activeBundleKeys(ctx, tx, snapshots[gamestrings.Domain], &versions)
	if err != nil {
		return gateReport{}, err
	}
	statics, err := passedStatics(ctx, tx)
	if err != nil {
		return gateReport{}, err
	}
	seasons, err := rating.ListSeasons(ctx, tx)
	if err != nil {
		return gateReport{}, err
	}
	templates, readable, err := h.gateTemplates(ctx, tx)
	if err != nil {
		return gateReport{}, err
	}
	campaigns, err := promoCampaigns(ctx, tx)
	if err != nil {
		return gateReport{}, err
	}
	if manifest, err := publishedManifest(snapshots[cosmetics.Domain]); err == nil {
		versions.ManifestVersion = manifest.ManifestVersion
	}
	result := contentgate.Run(contentgate.Inputs{
		LaunchAt: launchAt, Now: now, Snapshots: snapshots, Shop: h.config.Shop, Thresholds: thresholds, RankedSeasons: seasons,
		Templates: templates, TemplatesReadable: readable, Campaigns: campaigns, PromoCoinsLimit: h.config.Promo.CoinsApprovalThreshold,
		PassedStatics: statics, Bundles: bundles,
	})
	return gateReport{ID: uuid.NewString(), Environment: h.config.Environment, LaunchAt: launchAt, RanAt: now, Runner: p.account.ID,
		Versions: versions, Thresholds: thresholds, Result: result}, nil
}

func gateSnapshots(ctx context.Context, tx pgx.Tx) (gateVersions, map[string][]byte, error) {
	versions := gateVersions{Domains: map[string]int64{}, StringBundles: map[gamestrings.Language]int{}}
	active, err := snapshot.Newest(ctx, tx)
	if err != nil && !errors.Is(err, snapshot.ErrNone) {
		return gateVersions{}, nil, err
	}
	versions.ConfigVersion = active.Version
	snapshots := map[string][]byte{}
	for _, domain := range contentgate.Domains {
		version, raw, err := latestSnapshot(ctx, tx, domain)
		if err != nil {
			return gateVersions{}, nil, err
		}
		versions.Domains[domain] = version
		if raw != nil {
			snapshots[domain] = raw
		}
	}
	return versions, snapshots, nil
}

func gateThresholds(raw []byte, version int64) ([]contentgate.Threshold, error) {
	values := liveops.Values{}
	if raw != nil {
		var err error
		if values, err = liveops.Parse(raw); err != nil {
			return nil, err
		}
	}
	var thresholds []contentgate.Threshold
	for _, key := range liveops.Keys() {
		if !strings.HasPrefix(key.Name, contentgate.ThresholdPrefix) {
			continue
		}
		threshold := contentgate.Threshold{Key: key.Name, Value: key.Default, Default: key.Default}
		if value, set := values[key.Name]; set {
			threshold.Value, threshold.ConfigVersion = value, &version
		}
		thresholds = append(thresholds, threshold)
	}
	return thresholds, nil
}

func (h *Handler) activeBundleKeys(ctx context.Context, tx pgx.Tx, raw []byte, versions *gateVersions) (map[gamestrings.Language]map[string]bool, error) {
	keys := map[gamestrings.Language]map[string]bool{}
	if raw == nil {
		return keys, nil
	}
	localization, err := gamestrings.ParseLocalization(raw)
	if err != nil {
		return nil, err
	}
	for language, version := range localization.Bundles {
		published, err := h.strings.In(tx).Bundle(ctx, language, version)
		if err != nil {
			return nil, err
		}
		var bundle gamestrings.Bundle
		if err := json.Unmarshal(published.Content, &bundle); err != nil {
			return nil, err
		}
		keys[language] = make(map[string]bool, len(bundle.Strings))
		for key := range bundle.Strings {
			keys[language][key] = true
		}
		versions.StringBundles[language] = version
	}
	return keys, nil
}

func passedStatics(ctx context.Context, tx pgx.Tx) (map[string]bool, error) {
	return idSet(ctx, tx, `select distinct item_id from cosmetic_assets where slot = $1 and status = $2`, assetSlotStatic, assetPassed)
}

func promoCampaigns(ctx context.Context, tx pgx.Tx) (map[string]bool, error) {
	return idSet(ctx, tx, `select campaign_id from promo_codes union select campaign_id from promo_templates`)
}

func idSet(ctx context.Context, tx pgx.Tx, query string, args ...any) (map[string]bool, error) {
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set, nil
}

func (h *Handler) gateTemplates(ctx context.Context, tx pgx.Tx) ([]promo.Template, bool, error) {
	if h.promos == nil {
		return nil, false, nil
	}
	templates, err := h.promos.In(tx).Templates(ctx)
	return templates, err == nil, err
}

func storeGateReport(ctx context.Context, tx pgx.Tx, report gateReport) error {
	results := make([]gateCheckResult, 0, len(report.Checks))
	for _, check := range report.Checks {
		results = append(results, gateCheckResult{ID: check.ID, Result: check.Result})
	}
	_, err := tx.Exec(ctx, `insert into content_gate_reports (id, environment, launch_at, ran_at, runner, verdict, versions, results, record)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		report.ID, report.Environment, report.LaunchAt, report.RanAt, report.Runner, report.Verdict, marshal(report.Versions), marshal(results), marshal(report))
	return err
}

func (h *Handler) currentDomainVersions(ctx context.Context) (map[string]int64, error) {
	rows, err := h.pool.Query(ctx, `select domain, max(version) from config_snapshots where domain = any($1) group by domain`, contentgate.Domains)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	current := map[string]int64{}
	for rows.Next() {
		var domain string
		var version int64
		if err := rows.Scan(&domain, &version); err != nil {
			return nil, err
		}
		current[domain] = version
	}
	return current, rows.Err()
}

func (h *Handler) contentGateReports(w http.ResponseWriter, r *http.Request, p principal) {
	current, err := h.currentDomainVersions(r.Context())
	if err != nil {
		h.refuse(w, err)
		return
	}
	rows, err := h.pool.Query(r.Context(), `select id::text, environment, launch_at, ran_at, runner, verdict, versions, results
		from content_gate_reports order by ran_at desc, id desc limit $1`, ContentGateHistoryLimit)
	if err != nil {
		h.refuse(w, err)
		return
	}
	reports, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (gateSummary, error) {
		var summary gateSummary
		var versions gateVersions
		if err := row.Scan(&summary.ID, &summary.Environment, &summary.LaunchAt, &summary.RanAt, &summary.Runner, &summary.Verdict, &versions, &summary.Results); err != nil {
			return gateSummary{}, err
		}
		summary.LaunchAt, summary.RanAt, summary.Stale = summary.LaunchAt.UTC(), summary.RanAt.UTC(), versions.staleAgainst(current)
		return summary, nil
	})
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, map[string]any{"reports": reports})
}

func (h *Handler) storedGateReport(ctx context.Context, id string) (json.RawMessage, gateVersions, error) {
	if uuid.Validate(id) != nil {
		return nil, gateVersions{}, errReportUnknown
	}
	var record json.RawMessage
	var versions gateVersions
	err := h.pool.QueryRow(ctx, `select record, versions from content_gate_reports where id = $1`, id).Scan(&record, &versions)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, gateVersions{}, errReportUnknown
	}
	return record, versions, err
}

func (h *Handler) contentGateReport(w http.ResponseWriter, r *http.Request, p principal) {
	record, versions, err := h.storedGateReport(r.Context(), r.PathValue("id"))
	if err != nil {
		h.refuse(w, err)
		return
	}
	current, err := h.currentDomainVersions(r.Context())
	if err != nil {
		h.refuse(w, err)
		return
	}
	reply(w, http.StatusOK, gateReportView{Report: record, Stale: versions.staleAgainst(current)})
}

func (h *Handler) exportContentGateReport(w http.ResponseWriter, r *http.Request, p principal) {
	id := r.PathValue("id")
	record, _, err := h.storedGateReport(r.Context(), id)
	if err != nil {
		h.refuse(w, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="content-gate-`+id+`.json"`)
	reply(w, http.StatusOK, record)
}
