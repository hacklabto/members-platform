package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

// applicants live in the mysql database
var ApplyDB *sql.DB

var ErrDuplicateApplicant = errors.New("duplicate applicant")

type Applicant struct {
	Name                   string
	Username               string
	Nickname               string
	ContactEmail           string
	ListEmail              string
	BioReason              string
	Sponsor                sql.NullString
	SecondSponsor          sql.NullString
	Picture                sql.NullString
	TokenType              string
	Twitter                sql.NullString
	Facebook               sql.NullString
	HeardFrom              sql.NullString
	ProfileHash            sql.NullString
	Website                sql.NullString
	PreferredGenderPronoun sql.NullString
}

func ConnectApplyMySQL() error {
	log.Println("connecting to mysql")
	dsn := os.Getenv("APPLY_DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("missing APPLY_DATABASE_URL in environment")
	}
	my, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	ApplyDB = my
	return nil
}

func InsertApplicant(ctx context.Context, a Applicant) error {
	_, err := ApplyDB.ExecContext(ctx,
		`INSERT INTO applicants (
			name, username, nickname, contact_email, list_email, bio_reason, sponsor,
			second_sponsor, picture, token_type, twitter, facebook, heard_from, profile_hash,
			website, preferred_gender_pronoun
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.Name,
		a.Username,
		a.Nickname,
		a.ContactEmail,
		a.ListEmail,
		a.BioReason,
		a.Sponsor,
		a.SecondSponsor,
		a.Picture,
		a.TokenType,
		a.Twitter,
		a.Facebook,
		a.HeardFrom,
		a.ProfileHash,
		a.Website,
		a.PreferredGenderPronoun,
	)
	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) && myErr.Number == 1062 {
		return ErrDuplicateApplicant
	}
	return err
}

// returns the picture filename for an applicant's profile hash
func GetApplicantPicture(ctx context.Context, hash string) (string, error) {
	var picture sql.NullString
	err := ApplyDB.QueryRowContext(ctx,
		"SELECT picture FROM applicants WHERE profile_hash = ?", hash,
	).Scan(&picture)
	if err != nil {
		return "", err
	}
	return picture.String, nil
}
