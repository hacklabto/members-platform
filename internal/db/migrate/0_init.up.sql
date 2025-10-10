create table members (
    username text not null primary key,
    name text not null,
    picture text,
    picture_thumb text,
    join_date text,
    refer_to text,
    contact_info text,
    interests text,
    badges text[],
    board bool not null default false,
    sudoer bool not null default false
);
