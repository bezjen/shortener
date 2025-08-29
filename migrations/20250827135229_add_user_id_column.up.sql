alter table t_short_url add column user_id varchar(50);

create index idx_short_url_user_id on t_short_url (user_id);