drop index if exists idx_short_url_user_id;

alter table if exists t_short_url drop column user_id;