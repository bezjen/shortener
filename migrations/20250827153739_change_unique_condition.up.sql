drop index idx_short_url_original_url;

create unique index idx_short_url_user_id_original_url on t_short_url (user_id, original_url);