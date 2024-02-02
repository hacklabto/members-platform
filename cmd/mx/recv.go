package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/emersion/go-message"
	"github.com/emersion/go-smtp"

	jobs "members-platform/internal/jobs"
)

type recvBackend struct{}

func (recvBackend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{rcpt: []string{}}, nil
}

type session struct {
	mailfrom string
	rcpt     []string
}

// todo: clear session
func (s *session) Reset() {}

func (s *session) AuthPlain(username, password string) error {
	return smtp.ErrAuthUnsupported
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	s.mailfrom = from
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.rcpt = append(s.rcpt, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read email data: %w", err)
	}

	_, err = message.Read(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("parse message: %w", err)
	}

	return jobs.EnqueueJob(jobs.JOB_MX_INBOUND, jobs.MxInboundJobData{
		MailFrom: s.mailfrom,
		Rcpt:     s.rcpt,
		Data:     data,
	})
}

func (s *session) Logout() error {
	return smtp.ErrAuthUnsupported
}
