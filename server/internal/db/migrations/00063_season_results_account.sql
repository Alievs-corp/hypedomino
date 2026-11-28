-- +goose Up
create index season_results_account on season_results (account_id);

-- +goose Down
drop index season_results_account;
