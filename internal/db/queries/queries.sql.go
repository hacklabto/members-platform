// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: queries.sql

package queries

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

const getMemberFull = `-- name: GetMemberFull :one
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
where id = $1
`

func (q *Queries) GetMemberFull(ctx context.Context, id int32) (Member, error) {
	row := q.db.QueryRowContext(ctx, getMemberFull, id)
	var i Member
	err := row.Scan(
		&i.ID,
		&i.PreferredName,
		&i.PreferredPronouns,
		&i.Username,
		&i.ContactEmail,
		&i.ListEmail,
		&i.PictureUrl,
		&i.BioFreeform,
		pq.Array(&i.UserGroups),
		&i.IsCurrentMember,
		&i.ApplicationID,
		&i.LegalName,
		&i.WaiverSignDate,
		&i.AccessCardID,
		&i.EmergencyContact,
		&i.HelcimSubscriptionID,
	)
	return i, err
}

const getMemberSanitized = `-- name: GetMemberSanitized :one
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
where id = $1
`

type GetMemberSanitizedRow struct {
	ID                int32
	PreferredName     string
	PreferredPronouns string
	Username          string
	ContactEmail      string
	ListEmail         sql.NullString
	PictureUrl        string
	BioFreeform       sql.NullString
	UserGroups        []string
	IsCurrentMember   bool
}

func (q *Queries) GetMemberSanitized(ctx context.Context, id int32) (GetMemberSanitizedRow, error) {
	row := q.db.QueryRowContext(ctx, getMemberSanitized, id)
	var i GetMemberSanitizedRow
	err := row.Scan(
		&i.ID,
		&i.PreferredName,
		&i.PreferredPronouns,
		&i.Username,
		&i.ContactEmail,
		&i.ListEmail,
		&i.PictureUrl,
		&i.BioFreeform,
		pq.Array(&i.UserGroups),
		&i.IsCurrentMember,
	)
	return i, err
}
