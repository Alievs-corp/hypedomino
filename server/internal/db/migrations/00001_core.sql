-- +goose Up
create table accounts (
    id uuid primary key default gen_random_uuid(),
    state text not null default 'active' check (state in ('active', 'deleting', 'deleted', 'banned')),
    country char(2),
    created_at timestamptz not null default now(),
    deletion_requested_at timestamptz,
    deleted_at timestamptz
);

create table auth_identities (
    provider text not null check (provider in ('google', 'apple')),
    subject text not null,
    account_id uuid not null references accounts (id),
    email text,
    created_at timestamptz not null default now(),
    primary key (provider, subject)
);
create index auth_identities_account_idx on auth_identities (account_id);

create table profiles (
    account_id uuid primary key references accounts (id),
    nickname text not null,
    nickname_changed_at timestamptz,
    level integer not null default 1 check (level >= 1),
    xp bigint not null default 0 check (xp >= 0),
    loadout jsonb not null default '{}'::jsonb,
    stats jsonb not null default '{}'::jsonb,
    updated_at timestamptz not null default now()
);
create unique index profiles_nickname_idx on profiles (lower(nickname));

create table inventory_items (
    account_id uuid not null references accounts (id),
    item_id text not null,
    source_type text not null,
    source_ref text not null,
    granted_at timestamptz not null default now(),
    revoked_at timestamptz,
    primary key (account_id, item_id)
);

create table wallets (
    account_id uuid primary key references accounts (id),
    balance bigint not null default 0 check (balance >= 0),
    updated_at timestamptz not null default now()
);

create table ledger_entries (
    id bigint generated always as identity primary key,
    account_id uuid not null references accounts (id),
    amount bigint not null,
    reason text not null,
    source_type text not null,
    source_ref text not null,
    idempotency_key text not null,
    config_version integer not null,
    balance_after bigint not null check (balance_after >= 0),
    created_at timestamptz not null default now(),
    unique (account_id, idempotency_key)
);
create index ledger_entries_account_idx on ledger_entries (account_id, created_at desc);
create unique index ledger_entries_purchase_idx on ledger_entries (idempotency_key) where reason = 'purchase';

create table matches (
    id uuid not null,
    started_at timestamptz not null,
    ended_at timestamptz,
    mode text not null check (mode in ('1v1', '2v2')),
    kind text not null check (kind in ('ranked', 'casual', 'custom')),
    region text,
    config_version integer not null,
    seed bigint not null,
    hands_played integer not null default 0,
    score_a integer not null default 0,
    score_b integer not null default 0,
    winner_team smallint check (winner_team in (0, 1)),
    cancelled boolean not null default false,
    replay bytea,
    primary key (id, started_at)
) partition by range (started_at);

create table match_participants (
    match_id uuid not null,
    started_at timestamptz not null,
    seat smallint not null check (seat between 0 and 5),
    team smallint not null check (team in (0, 1)),
    account_id uuid references accounts (id),
    bot boolean not null default false,
    result text check (result in ('win', 'loss', 'cancelled')),
    primary key (match_id, started_at, seat),
    foreign key (match_id, started_at) references matches (id, started_at)
) partition by range (started_at);
create index match_participants_account_idx on match_participants (account_id, started_at desc);

create table matches_default partition of matches default;
create table match_participants_default partition of match_participants default;

-- +goose StatementBegin
create function match_partitions_ensure(from_month date, months integer) returns integer
language plpgsql as $$
declare
    month_start date := date_trunc('month', from_month)::date;
    month_end date;
    suffix text;
    created integer := 0;
begin
    for i in 1 .. months loop
        month_end := (month_start + interval '1 month')::date;
        suffix := to_char(month_start, 'YYYYMM');
        if to_regclass('matches_' || suffix) is null then
            execute format(
                'create table %I partition of matches for values from (%L) to (%L)',
                'matches_' || suffix,
                month_start::timestamp at time zone 'UTC',
                month_end::timestamp at time zone 'UTC');
            execute format(
                'create table %I partition of match_participants for values from (%L) to (%L)',
                'match_participants_' || suffix,
                month_start::timestamp at time zone 'UTC',
                month_end::timestamp at time zone 'UTC');
            created := created + 1;
        end if;
        month_start := month_end;
    end loop;
    return created;
end
$$;
-- +goose StatementEnd

select match_partitions_ensure(current_date, 6);

-- +goose Down
drop function match_partitions_ensure(date, integer);
drop table match_participants;
drop table matches;
drop table ledger_entries;
drop table wallets;
drop table inventory_items;
drop table profiles;
drop table auth_identities;
drop table accounts;
