-- +goose Up
create table content_gate_reports (
    id uuid primary key,
    environment text not null check (environment in ('staging', 'production')),
    launch_at timestamptz not null,
    ran_at timestamptz not null,
    runner text not null,
    verdict text not null check (verdict in ('green', 'red')),
    versions jsonb not null,
    results jsonb not null,
    record jsonb not null
);
create index content_gate_reports_ran_idx on content_gate_reports (ran_at desc, id desc);

-- +goose StatementBegin
create function content_gate_reports_immutable() returns trigger language plpgsql as $$
begin
    raise exception 'content_gate_reports is append-only';
end;
$$;
-- +goose StatementEnd

create trigger content_gate_reports_immutable
    before update or delete or truncate on content_gate_reports
    for each statement execute function content_gate_reports_immutable();

-- +goose Down
drop table content_gate_reports;
drop function content_gate_reports_immutable();
