-- todo: moderator approval
create type list_permission as enum ('everyone', 'members', 'whitelist');

create table lists (
    id serial primary key,
    name text not null,
    send_permissions list_permission default 'members',
    read_permissions list_permission default 'everyone'
);

create table list_members (
    list_id int not null,
    member_email text not null,
    can_read bool default false,
    can_write bool default false,
    primary key (list_id, member_email)
);

create function can_send_to_list(listid int, email text) returns bool as $$
declare whitelisted bool;
declare perm list_permission;
begin
    whitelisted := can_read from list_members where list_id = listid and member_email = email;
    perm := send_permissions as perm from lists where id = listid;
    case perm
        when 'everyone' then return true;
        when 'members' then return whitelisted is not null;
        when 'whitelist' then return coalesce(whitelisted, false);
    end case;
end
$$ language plpgsql;

create table list_messages (
    -- full message is in /path/to/archive/listid/pkid.eml
    -- or s3, or whatever
    id uuid default gen_random_uuid(),
    message_id text unique not null,
    list_id int not null,
    parent_message_id text,
    thread_id uuid not null, -- coalesce(parent.thread_id, gen_random_uuid())
    mailfrom text not null,
    subject text not null,
    ts timestamp not null,
    primary key (id, list_id)
);
create index idx_list_messages_by_message_id on list_messages (message_id, list_id);
