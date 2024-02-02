package jobs

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"members-platform/internal/db"
	"members-platform/internal/listdb"
	"members-platform/internal/listdb/queries"
	"strings"

	"github.com/emersion/go-message"
	"github.com/taylorchu/work"
)

var JOB_MX_INBOUND JobName = "mx-inbound"

type MxInboundJobData struct {
	MailFrom string
	Rcpt     []string
	Data     []byte
}

func RunMXInbound(ctx context.Context, j *work.Job, do *work.DequeueOptions) error {
	var data MxInboundJobData
	if err := json.Unmarshal(j.Payload, &data); err != nil {
		return fmt.Errorf("unmarshal job data: %w", err)
	}

	mail, err := message.Read(bytes.NewReader(data.Data))
	if err != nil {
		return fmt.Errorf("read message: %w", err)
	}

	if ok, err := validateMail(ctx, mail, data.MailFrom, data.Rcpt); err != nil {
		return fmt.Errorf("validate email: %w", err)
	} else if !ok {
		// validation succeeded but email is not ok
		// the validateMail function already notified the user, so let's drop this job here
		return nil
	}

	// the email is ok and we can deliver it (or execute a command or something)

	q := work.NewRedisQueue(db.Redis)

	for _, addr := range data.Rcpt {
		typ, err := getEmailType(ctx, addr)
		if err != nil {
			return fmt.Errorf("failed to get email type for %s: %w", addr, err)
		}
		switch typ {
		case emailtype_deliver:
			EnqueueJob(JOB_MX_DELIVER, MxDeliverJobData{
				ListAddr: addr,
				Message:  string(data.Data),
			})
		case emailtype_command:
			// queue command job
		}

	}

	return nil
}

// headers we add to emails (cannot be contained in received emails)
var restrictedHeaders = []string{
	"List-Archive",
	"List-Help",
	"List-ID",
	"List-Owner",
	"List-Post",
	"List-Subscribe",
	"List-Unsubscribe",
	"Reply-To", // this is always the list address
	"Return-Receipt-To",
	"Disposition-Notification-To",
}

var requiredHeaders = []string{
	"From",
	"To",
	"Subject",
	"Message-ID",
}

func validateMail(ctx context.Context, m *message.Entity, mailfrom string, rcpt []string) (bool, error) {
	fields := m.Header.FieldsByKey("X-BeenThere")
	for {
		if !fields.Next() {
			break
		}
		if strings.ToLower(fields.Key()) == "x-beenthere" && strings.Contains(fields.Value(), "lists.hacklab.to") {
			// we already saw this message, drop it
			log.Println("already saw this message!")
			return false, nil
		}
	}

	errors := []string{}

	// no list addresses found
	// sending to a list and a command address at the same time
	{
		var hasDeliverAddress bool
		var hasCommandAddress bool

		for _, addr := range rcpt {
			typ, err := getEmailType(ctx, addr)
			if err != nil {
				return false, fmt.Errorf("failed to get email type for %s: %w", addr, err)
			}
			switch typ {
			case emailtype_deliver:
				hasDeliverAddress = true
			case emailtype_command:
				hasCommandAddress = true
				// if we found unknown addresses, that's fine
			}
		}

		if hasDeliverAddress && hasCommandAddress {
			errors = append(errors, "Cannot send email to list and command address at the same time")
		}
		if !hasDeliverAddress && !hasCommandAddress {
			errors = append(errors, "No known list addresses found")
		}
	}

	// restricted headers
	// required headers not found (from, to, subject, message-id)
	{
		for _, header := range restrictedHeaders {
			if m.Header.Get(header) != "" {
				errors = append(errors, fmt.Sprintf("Header %s is forbidden", header))
			}
		}
		for _, header := range requiredHeaders {
			if m.Header.Get(header) == "" {
				errors = append(errors, fmt.Sprintf("Header %s is required", header))
			}
		}
	}

	// not allowed to send to list
	for _, addr := range rcpt {
		// todo: don't copy this from getEmailType
		parts := strings.Split(addr, "@")
		if len(parts) != 2 || parts[1] != "hacklab.to" {
			// lists addresses don't contain >1 '@'
			continue
		}
		user_parts := strings.Split(parts[0], "+")
		if len(user_parts) > 1 {
			// lists addresses don't contain >1 '+'
			// also we don't care about command addresses
			continue
		}
		ok, err := listdb.DB.CheckCanSendToList(ctx, queries.CheckCanSendToListParams{
			Name:  user_parts[0],
			Email: mailfrom,
		})
		if err != nil {
			return false, fmt.Errorf("failed to check if %s can send to list %s: %w", mailfrom, user_parts[0], err)
		}
		if !ok {
			errors = append(errors, fmt.Sprintf("Not allowed to send emails to list '%s'", addr))
		}
	}

	// todo
	// no text/plain or text/html part
	// warn on no text/plain part or has text/html part (only once)

	if len(errors) > 0 {
		buf := bytes.NewBuffer([]byte(""))
		if err := m.WriteTo(buf); err != nil {
			return false, fmt.Errorf("write original email: %w", err)
		}

		if err := EnqueueJob(JOB_MX_REJECT, MxRejectJobData{
			Dest:         mailfrom,
			Errors:       errors,
			OriginalMail: buf.String(),
		}); err != nil {
			return false, fmt.Errorf("failed to add to reject job queue: %w", err)
		}
	}

	return true, nil
}

type emailtype int

const (
	emailtype_deliver emailtype = iota
	emailtype_command
	emailtype_unknown
)

func getEmailType(ctx context.Context, email string) (emailtype, error) {
	// this is really basic and will probably break at some point
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] != "hacklab.to" {
		// lists addresses don't contain >1 '@'
		return emailtype_unknown, nil
	}
	user_parts := strings.Split(parts[0], "+")
	if len(user_parts) > 2 {
		// lists addresses don't contain >1 '+'
		return emailtype_unknown, nil
	}
	_, err := listdb.DB.GetListByEmail(ctx, user_parts[0]+"@"+parts[1])
	switch {
	case err == sql.ErrNoRows:
		return emailtype_unknown, nil
	case err != nil:
		return emailtype_unknown, err
	}
	if len(user_parts) > 1 {
		switch user_parts[1] {
		case "subscribe", "unsubscribe":
			return emailtype_command, nil
		}
		return emailtype_unknown, nil
	}
	return emailtype_deliver, nil
}
