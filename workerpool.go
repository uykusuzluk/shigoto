package shigoto

import (
	"encoding/json"
	"log"
	"time"
)

type workerFn func(Job, *log.Logger, *workerPool)

type wpState int

const (
	wpInit wpState = iota
	wpRunning
	wpClosing
)

type workerPool struct {
	count int
	pool  chan workerFn
	state wpState
}

func newWorkerPool(n int) *workerPool {
	wp := &workerPool{
		count: n,
		pool:  make(chan workerFn, n),
		state: wpInit,
	}

	for i := 0; i < n; i++ {
		wp.pool <- work
	}
	wp.state = wpRunning
	return wp
}

func (wp *workerPool) get() workerFn {
	return <-wp.pool
}

func (wp *workerPool) put() {
	if wp.state == wpClosing {
		return
	}
	// FIXME: !!! Will panic if channel is closed right here. Recover?
	wp.pool <- work
}

func (wp *workerPool) close() {
	wp.state = wpClosing
	close(wp.pool)
}

func work(job Job, l *log.Logger, wp *workerPool) {
	defer wp.put()
	l.Println("worker work: job received...")

	correct, err := job.checkPayload()
	if !correct {
		l.Println(err.Error())
		return
	}

	stale, err := job.Expired()
	if stale {
		l.Println(err.Error())
		return
	}

	l.Println("worker work: job ID: ", job.UUID, " will be serialized and run.")
	job.Attemps++
	job.StartedAt = time.Now().UTC()

	objBlueprint := jobContainer[job.PayloadType]
	newObj, err := objBlueprint.New()
	if err != nil {
		l.Println("worker work: cannot create a new variable from blueprint: Payload: ", job.Payload, " Type: ", job.PayloadType)
	}

	err = json.Unmarshal(job.Payload, &newObj)
	if err != nil {
		l.Println("worker work: cannot unmarshal object. Payload: ", job.Payload, " Type: ", job.PayloadType)
	}

	err = newObj.Run()
	if err != nil {
		l.Printf("Job with ID: %s has failed. Job: %+v", job.UUID, job)
		// TODO: Fail job stuff (check attemps left, requeue, make failed_jobs db etc...)
	}

	l.Println("worker work: job ran successfully.")
}
