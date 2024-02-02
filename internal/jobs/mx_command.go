package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/taylorchu/work"
)

var JOB_MX_COMMAND JobName = "mx-command"

type MxCommandJobData struct {
}

func RunMXCommand(ctx context.Context, j *work.Job, do *work.DequeueOptions) error {
	var data MxCommandJobData
	if err := json.Unmarshal(j.Payload, &data); err != nil {
		return fmt.Errorf("unmarshal job data: %w", err)
	}
	return nil
}
