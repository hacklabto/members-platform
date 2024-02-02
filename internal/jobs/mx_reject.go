package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"members-platform/internal/mailer"
	"strings"

	"github.com/emersion/go-message"
	"github.com/taylorchu/work"
)

var JOB_MX_REJECT JobName = "mx-reject"

type MxRejectJobData struct {
	Dest         string
	Errors       []string
	OriginalMail string
}

func buildRejectionMessage(errors []string) *bytes.Buffer {
	errString := strings.Join(errors, "\n- ")
	return bytes.NewBufferString(strings.TrimSpace(`
Hello, this is an automated message from the list server at hacklab.to.
We received your email, but were unable to process it.

` + errString + `

If you have any questions about the above or require more assistance, just
reply to this email to reach Operations.
`))
}

func RunMXReject(ctx context.Context, j *work.Job, do *work.DequeueOptions) error {
	var data MxRejectJobData
	if err := json.Unmarshal(j.Payload, &data); err != nil {
		return fmt.Errorf("unmarshal job data: %w", err)
	}

	originalMessage, err := message.New(
		message.HeaderFromMap(map[string][]string{
			"Content-Type":        {"message/rfc822"},
			"Content-Disposition": {"attachment; filename=\"Email.eml\""},
		}),
		bytes.NewBufferString(data.OriginalMail),
	)
	if err != nil {
		return fmt.Errorf("create entity for original message: %w", err)
	}

	newMessage, err := message.New(
		message.HeaderFromMap(map[string][]string{
			"Content-Type": {"text/plain; charset=us-ascii"},
		}),
		buildRejectionMessage(data.Errors),
	)
	if err != nil {
		return fmt.Errorf("create entity for new message: %w", err)
	}

	multipartMessage, err := message.NewMultipart(
		message.HeaderFromMap(map[string][]string{
			"From":     {"lists.hacklab.to <operations+automated@hacklab.to>"},
			"Reply-To": {"operations@hacklab.to"},
			"To":       {data.Dest},
			"Subject":  {"Re: Your recent email to lists.hacklab.to"},
		}),
		[]*message.Entity{newMessage, originalMessage},
	)
	if err != nil {
		return fmt.Errorf("create multipart message: %w", err)
	}

	bs := bytes.NewBufferString("")
	if err := multipartMessage.WriteTo(bs); err != nil {
		return fmt.Errorf("write multipart message: %w", err)
	}

	return mailer.DoSendEmail(data.Dest, bs.String())
}
