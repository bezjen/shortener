create table t_short_url(
    short_url varchar(8) not null,
    original_url varchar(255) not null,
    primary key (short_url)
);

create unique index idx_short_url_original_url on t_short_url (original_url);