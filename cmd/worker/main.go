package main

import (
	"log"
	"members-platform/internal/db"
	"members-platform/internal/jobs"
	"members-platform/internal/listdb"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/taylorchu/work"
	"github.com/taylorchu/work/middleware/discard"
	"github.com/taylorchu/work/middleware/logrus"
)

func main() {
	if err := db.ConnectPG(false); err != nil {
		log.Fatalln(err)
	}
	if err := db.ConnectRedis(); err != nil {
		log.Fatalln(err)
	}
	if err := listdb.ConnectPG(false); err != nil {
		log.Fatalln(err)
	}

	w := work.NewWorker(&work.WorkerOptions{
		Queue: work.NewRedisQueue(db.Redis),
	})

	// or something
	jobOpts := &work.JobOptions{
		MaxExecutionTime: time.Minute,
		IdleWait:         4 * time.Second,
		NumGoroutines:    4,
		HandleMiddleware: []work.HandleMiddleware{
			logrus.HandleFuncLogger,
			discard.After(time.Hour),
		},
	}

	w.RegisterWithContext(string(jobs.JOB_MX_INBOUND), jobs.RunMXInbound, jobOpts)
	w.RegisterWithContext(string(jobs.JOB_MX_COMMAND), jobs.RunMXCommand, jobOpts)
	w.RegisterWithContext(string(jobs.JOB_MX_DELIVER), jobs.RunMXDeliver, jobOpts)
	w.RegisterWithContext(string(jobs.JOB_MX_REJECT), jobs.RunMXReject, jobOpts)

	w.Start()
	defer w.Stop()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
