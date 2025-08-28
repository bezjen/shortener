drop index if exists idx_short_url_is_deleted;

alter table if exists t_short_url drop column is_deleted;