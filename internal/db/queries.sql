-- name: GetMembers :many
select * from members order by lower(username) asc;

-- name: GetMemberByUsername :one
select * from members where username = $1;

-- name: UpdateProfile :exec
update members set refer_to = $1, contact_info = $2, interests = $3 where username = $4;

-- name: Memberize :exec
insert into members
(username, name, picture, picture_thumb, join_date, contact_info)
values
($1, $2, $3, $4, $5, $6);
