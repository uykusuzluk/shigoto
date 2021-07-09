package shigoto

import (
	"encoding/json"
	"log"
	"time"
)

// Worker is the responsible for reading jobs sent by the Listener and starting their Run() methods
type Worker struct {
	log       *log.Logger
	jobsToRun <-chan Job
	stopChan  chan struct{}
}

// NewWorker is the pseudo-constructor of Worker struct
func newWorker(jobChan <-chan Job, l *log.Logger) *Worker {
	w := &Worker{log: l}
	w.initialize(jobChan)
	return w
}

func (w *Worker) initialize(jobChan <-chan Job) {
	w.jobsToRun = jobChan
	w.stopChan = make(chan struct{})
}

// Work is the main method of the Worker which will run concurrently
func (w *Worker) work() {
	for {
		select {
		case <-w.stopChan:
			w.log.Println("worker work: received stop message and returning...")
			return
		case job := <-w.jobsToRun:
			w.log.Println("worker work: job received...")

			correct, err := job.checkPayload()
			if !correct {
				w.log.Println(err.Error())
				continue
			}

			stale, err := job.Expired()
			if stale {
				w.log.Println(err.Error())
				continue
			}

			w.log.Println("worker work: job ID: ", job.UUID, " will be serialized and run.")
			job.Attemps++
			job.StartedAt = time.Now().UTC()

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
			}

			w.log.Println("worker work: job ran successfully.")
		}
	}
}

// Stop method is called to kill a running Worker
func (w *Worker) stop() {
	w.jobsToRun = nil
	w.stopChan <- struct{}{}
}
