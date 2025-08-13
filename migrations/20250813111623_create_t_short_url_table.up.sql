create table t_short_url(
    id uuid not null,
    short_url varchar(8) not null,
    original_url varchar(255) not null,
    primary key (id)
);

create index idx_short_url_short_url on t_short_url(short_url);