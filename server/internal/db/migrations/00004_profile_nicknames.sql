-- +goose Up
drop index profiles_nickname_idx;
alter table profiles add column nickname_key text not null;
create unique index profiles_nickname_key_idx on profiles (nickname_key);

create table nickname_history (
    id bigint generated always as identity primary key,
    account_id uuid not null references accounts (id),
    nickname text not null,
    nickname_key text not null,
    replaced_at timestamptz not null default now()
);
create index nickname_history_key_idx on nickname_history (nickname_key, replaced_at desc);
create index nickname_history_account_idx on nickname_history (account_id, replaced_at desc);

-- +goose Down
drop table nickname_history;
drop index profiles_nickname_key_idx;
alter table profiles drop column nickname_key;
create unique index profiles_nickname_idx on profiles (lower(nickname));
