alter table t_short_url add column is_deleted boolean not null default false;

create index idx_short_url_is_deleted on t_short_url (is_deleted);