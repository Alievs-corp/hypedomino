-- +goose Up
alter table match_participants add column loadout jsonb not null default '{}'::jsonb;

-- +goose Down
alter table match_participants drop column loadout;
