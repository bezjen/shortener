create table t_short_url(
    short_url varchar(8) not null,
    original_url varchar(255) not null,
    primary key (short_url)
);