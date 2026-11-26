-- +goose Up
alter table promo_templates add column launch_role text check (launch_role in ('welcome', 'creator', 'compensation'));
create unique index promo_templates_launch_role on promo_templates (launch_role) where launch_role is not null;

-- +goose Down
drop index promo_templates_launch_role;
alter table promo_templates drop column launch_role;
