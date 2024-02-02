-- name: GetListByEmail :one
select id, name from lists where name || '@hacklab.to' = $1;

-- name: GetListRecipients :many
select member_email from list_members where list_id = $1;

-- name: GetListById :one
select * from lists where id = $1;

-- name: CheckCanSendToList :one
select can_send_to_list((select list_id from lists where name = $1), $2);

-- name: AddMessage :exec
insert into list_messages (
    message_id,
    thread_id,
    mailfrom,
    subject,
    ts
) values (
    $1,
    coalesce((select thread_id from list_messages m where m.message_id = $2), gen_random_uuid()),
    $3,
    $4,
    $5
);
