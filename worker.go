package shigoto

import (
	"encoding/json"
	"log"
)

// Worker is the responsible for reading jobs sent by the Listener and starting their Run() methods
type worker struct {
	log      *log.Logger
	jobchan  <-chan Job
	stopChan chan struct{}
}

// NewWorker is the pseudo-constructor of Worker struct
func newWorker(jobChan <-chan Job, l *log.Logger) *worker {
	return &worker{
		log:      l,
		jobchan:  jobChan,
		stopChan: make(chan struct{}),
	}
}

// Work is the main method of the Worker which will run concurrently
func (w *worker) work() {
	defer w.close()
	for {
		select {
		case <-w.stopChan:
			w.log.Println("worker work: received stop message and returning...")
			return
		case job := <-w.jobchan:
			w.log.Println("worker work: job received...")

			err := job.checkPayload()
			if err != nil {
				w.log.Println(err.Error())
				continue
			}

			err = job.expired()
			if err != nil {
				w.log.Println(err.Error())
				continue
			}

			w.log.Println("worker work: job ID: ", job.UUID, " will be serialized and run.")
			job.start()

			objBlueprint := jobContainer[job.PayloadType]
			newObj, err := objBlueprint.New()
			if err != nil {
				w.log.Println("worker work: cannot create a new variable from blueprint: Payload: ", job.Payload, " Type: ", job.PayloadType)
			}

			err = json.Unmarshal(job.Payload, &newObj)
			if err != nil {
				w.log.Println("worker work: cannot unmarshal object. Payload: ", job.Payload, " Type: ", job.PayloadType)
			}

			err = newObj.Run()
			if err != nil {
				w.log.Printf("Job with ID: %s has failed. Job: %+v", job.UUID, job)
				// TODO: Fail job stuff (check attemps left, requeue, make failed_jobs db etc...)
				// TODO: If attempts > tries requeue; else fail -> log, db, taskboard reporting
			}

			w.log.Println("worker work: job ran successfully.")
		}
	}
}

// Stop method is called to kill a running Worker
func (w *worker) stop() {
	w.jobchan = nil
	w.stopChan <- struct{}{}
}

// Close method for cleaning up a Worker
func (w *worker) close() {
	close(w.stopChan)
}
