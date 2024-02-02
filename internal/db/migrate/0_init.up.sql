-- todo: proper schema idk

create table temp_waiver_signatures (
    signature_id uuid not null,
    member_username text,
    legal_name text not null,
    preferred_name text,
    date text not null,
    emergency_contact_name text not null,
    emergency_contact_number text not null,
    signature text not null,
    witness_member text
);

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
    join_date timestamp not null,

    -- board-only fields
    application_id int not null references applicants(id),
    legal_name text,
    waiver_sign_date timestamp,
    access_card_id text,
    emergency_contact text,
    balance_owing int,
    helcim_subscription_id text
);

create table bins (
    id text primary key,
    assigned_to_member int references members(id) on delete set null,
    assigned_to_function text,
    check (assigned_to_member is null or assigned_to_function is null)
);
