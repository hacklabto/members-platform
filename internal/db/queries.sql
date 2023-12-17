-- name: GetMemberSanitized :one
select
    id,
    preferred_name,
    preferred_pronouns,
    username,
    contact_email,
    list_email,
    picture_url,
    bio_freeform,
    user_groups,
    is_current_member
from members
where id = $1;

-- name: GetMemberFull :one
select
    id,
    preferred_name,
    preferred_pronouns,
    username,
    contact_email,
    list_email,
    picture_url,
    bio_freeform,
    user_groups,
    is_current_member,
    application_id,
    legal_name,
    waiver_sign_date,
    access_card_id,
    emergency_contact,
    helcim_subscription_id
from members
where id = $1;
