// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0

package queries

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Applicant struct {
	ID                int32
	PreferredName     string
	PreferredPronouns sql.NullString
	Nickname          string
	Username          string
	ContactEmail      string
	ListEmail         sql.NullString
	ApplicationReason string
	Sponsor1          string
	Sponsor2          string
	PictureUrl        string
	HeardFrom         sql.NullString
	Links             sql.NullString
	TokenType         int32
	ApplicationState  int32
}

type Bin struct {
	ID                 string
	AssignedToMember   sql.NullInt32
	AssignedToFunction sql.NullString
}

type Member struct {
	ID                   int32
	PreferredName        string
	PreferredPronouns    string
	Username             string
	ContactEmail         string
	ListEmail            sql.NullString
	PictureUrl           string
	BioFreeform          sql.NullString
	UserGroups           []string
	IsCurrentMember      bool
	JoinDate             time.Time
	ApplicationID        int32
	LegalName            sql.NullString
	WaiverSignDate       sql.NullTime
	AccessCardID         sql.NullString
	EmergencyContact     sql.NullString
	BalanceOwing         sql.NullInt32
	HelcimSubscriptionID sql.NullString
}

type TempWaiverSignature struct {
	SignatureID            uuid.UUID
	MemberUsername         sql.NullString
	LegalName              string
	PreferredName          sql.NullString
	Date                   string
	EmergencyContactName   string
	EmergencyContactNumber string
	Signature              string
	WitnessMember          sql.NullString
}
