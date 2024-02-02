package jobs

import (
	"encoding/json"
	"fmt"
	"members-platform/internal/db"

	"github.com/taylorchu/work"
)

type JobName string

func EnqueueJob(name JobName, data any) error {
	var err error
	job := work.NewJob()

	job.Payload, err = json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal job data: %w", err)
	}

	return work.NewRedisQueue(db.Redis).Enqueue(job, &work.EnqueueOptions{QueueID: string(name)})
}
