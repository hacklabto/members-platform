package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"members-platform/internal/listdb"
	"os"

	"github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
	"github.com/emersion/go-smtp"
	"github.com/taylorchu/work"
)

var JOB_MX_DELIVER JobName = "mx-deliver"

type MxDeliverJobData struct {
	ListAddr string
	Message  string
}

func RunMXDeliver(ctx context.Context, j *work.Job, do *work.DequeueOptions) error {
	var data MxDeliverJobData
	if err := json.Unmarshal(j.Payload, &data); err != nil {
		return fmt.Errorf("unmarshal job data: %w", err)
	}

	msg, err := message.Read(bytes.NewReader([]byte(data.Message)))
	if err != nil {
		return fmt.Errorf("parse message: %w", err)
	}

	ml, err := listdb.DB.GetListByEmail(ctx, data.ListAddr)
	if err != nil {
		return fmt.Errorf("get list: %w", err)
	}

	// add list headers
	// https://www.rfc-editor.org/rfc/rfc4021#section-2.1
	{
		msg.Header.Add("List-Archive", fmt.Sprintf("https://lists.hacklab.to/archives/%s/", ml.Name))
		msg.Header.Add("List-Help", fmt.Sprintf("<mailto:operations@hacklab.to?subject:list-%s-help>", ml.Name))
		msg.Header.Add("List-ID", fmt.Sprintf("%s <%s.lists.hacklab.to>", ml.Name, ml.Name))
		msg.Header.Add("List-Owner", "<mailto:operations@hacklab.to>")
		msg.Header.Add("List-Post", fmt.Sprintf("<mailto:%s@hacklab.to>", ml.Name))
		msg.Header.Add("List-Subscribe", fmt.Sprintf("<mailto:%s+subscribe@hacklab.to>", ml.Name))
		msg.Header.Add("List-Unsubscribe", fmt.Sprintf("<mailto:%s+unsubscribe@hacklab.to>", ml.Name))

		msg.Header.Add("X-BeenThere", "lists.hacklab.to")
	}

	rcpts, err := listdb.DB.GetListRecipients(context.Background(), ml.ID)
	if err != nil {
		return fmt.Errorf("get list recipients: %w", err)
	}

	smtpServer := os.Getenv("SMTP_URL")
	if smtpServer == "" {
		return fmt.Errorf("missing SMTP_URL in environment")
	}

	c, err := smtp.Dial(smtpServer)
	if err != nil {
		return err
	}
	defer c.Quit()

	if err = c.Hello("lists.hacklab.to"); err != nil {
		return err
	}

	if err = c.Mail("list-bounces@hacklab.to", nil); err != nil {
		return err
	}

	for _, addr := range rcpts {
		log.Println("send to", addr)
		if err = c.Rcpt(addr, nil); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	if err := msg.WriteTo(w); err != nil {
		return err
	}

	return w.Close()
}

func saveMailToFS(m message.Header, body io.Reader) error {
	textproto.WriteHeader(os.Stdout, m.Header)
	io.Copy(os.Stdout, body)
	return nil
}

func parseMailForFTS(*message.Entity) string {
	// todo
	return ""
}
