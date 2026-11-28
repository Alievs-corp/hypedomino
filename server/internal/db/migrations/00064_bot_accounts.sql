-- +goose Up
alter table accounts add column is_bot boolean not null default false;
alter table auth_identities drop constraint auth_identities_provider_check;
alter table auth_identities add constraint auth_identities_provider_check check (provider in ('google', 'apple', 'bot'));

-- +goose Down
delete from auth_identities where provider = 'bot';
alter table auth_identities drop constraint auth_identities_provider_check;
alter table auth_identities add constraint auth_identities_provider_check check (provider in ('google', 'apple'));
alter table accounts drop column is_bot;
