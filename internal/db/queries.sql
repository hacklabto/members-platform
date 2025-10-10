-- name: GetMembers :many
select * from members order by lower(username) asc;
