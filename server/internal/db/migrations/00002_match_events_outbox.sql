-- +goose Up
create table match_events (
    match_id uuid not null,
    started_at timestamptz not null,
    seq integer not null check (seq >= 0),
    kind text not null,
    seat smallint check (seat between 0 and 5),
    payload jsonb not null default '{}'::jsonb,
    at timestamptz not null,
    primary key (match_id, started_at, seq),
    foreign key (match_id, started_at) references matches (id, started_at)
) partition by range (started_at);

create table match_events_default partition of match_events default;

create table outbox (
    id bigint generated always as identity primary key,
    topic text not null,
    key text not null,
    payload jsonb not null,
    created_at timestamptz not null default now(),
    published_at timestamptz,
    attempts integer not null default 0,
    unique (topic, key)
);
create index outbox_pending_idx on outbox (id) where published_at is null;

-- +goose StatementBegin
create or replace function match_partitions_ensure(from_month date, months integer) returns integer
language plpgsql as $$
declare
    month_start date := date_trunc('month', from_month)::date;
    month_end date;
    suffix text;
    created integer := 0;
    parent text;
begin
    for i in 1 .. months loop
        month_end := (month_start + interval '1 month')::date;
        suffix := to_char(month_start, 'YYYYMM');
        foreach parent in array array['matches', 'match_participants', 'match_events'] loop
            if to_regclass(parent || '_' || suffix) is null then
                execute format(
                    'create table %I partition of %I for values from (%L) to (%L)',
                    parent || '_' || suffix,
                    parent,
                    month_start::timestamp at time zone 'UTC',
                    month_end::timestamp at time zone 'UTC');
                created := created + 1;
            end if;
        end loop;
        month_start := month_end;
    end loop;
    return created;
end
$$;
-- +goose StatementEnd

select match_partitions_ensure(current_date, 6);

-- +goose Down
drop table outbox;
drop table match_events;

-- +goose StatementBegin
create or replace function match_partitions_ensure(from_month date, months integer) returns integer
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
