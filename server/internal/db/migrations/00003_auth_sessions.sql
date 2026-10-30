-- +goose Up
create table auth_sessions (
    id uuid primary key default gen_random_uuid(),
    account_id uuid not null references accounts (id),
    device_id text not null,
    refresh_hash bytea not null,
    previous_hash bytea,
    rotated_at timestamptz,
    generation integer not null default 1 check (generation >= 1),
    created_at timestamptz not null default now(),
    last_seen_at timestamptz not null default now(),
    expires_at timestamptz not null,
    revoked_at timestamptz,
    revoke_reason text check (revoke_reason in ('logout', 'logout_all', 'reuse', 'device', 'expired'))
);
create index auth_sessions_account_idx on auth_sessions (account_id) where revoked_at is null;

-- +goose Down
drop table auth_sessions;
