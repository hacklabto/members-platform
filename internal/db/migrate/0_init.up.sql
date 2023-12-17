-- todo: proper schema idk

create table applicants (
    id serial primary key,
    preferred_name text not null,
    preferred_pronouns text,
    nickname text not null,
    username text unique not null,
    contact_email text unique not null,
    list_email text,
    application_reason text not null,
    sponsor1 text not null,
    sponsor2 text not null,
    picture_url text unique not null,
    heard_from text,
    links text,
    token_type int not null,
    application_state int not null default 1
);

create table members (
    id serial primary key,
    preferred_name text not null,
    preferred_pronouns text not null,
    username text unique not null,
    contact_email text unique not null,
    list_email text,
    picture_url text unique not null,
    bio_freeform text,
    user_groups text[] not null default array[]::text[],
    is_current_member bool not null default true,

    -- board-only fields
    application_id int not null references applicants(id),
    legal_name text,
    waiver_sign_date timestamp,
    access_card_id text,
    emergency_contact text,
    helcim_subscription_id text
);
