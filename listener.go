package shigoto

import (
	"encoding/json"
	"fmt"
	"log"
)

// Listener is the single point of loading the channel with jobs from a queue source,
// to send them to the registered active workers
type listener struct {
	numWorkers int
	qName      string

	log       *log.Logger
	taskboard TaskBoard

	workers []*Worker

	// TODO: turn it into a send-only channel?
	jobsToRun chan Job
	stopChan  chan struct{}
}

func newListener(queue string, workers int, s *Shigoto) (*listener, error) {
	l := &listener{
		log:        s.log,
		taskboard:  s.taskBoard,
		qName:      queue,
		numWorkers: workers,
	}
	l.init()
	return l, nil
}

func (l *listener) init() {
	jobChan := make(chan Job, l.numWorkers*2)
	l.jobsToRun = jobChan
	for i := 0; i < l.numWorkers; i++ {
		w := NewWorker(jobChan, l.log)
		l.workers = append(l.workers, w)
		go w.work()
	}
}

func (l *listener) listen() {
	for {
		select {
		case <-l.stopChan:
			l.stopWorkers()
			return
		default:
			job := Job{}
			jobJSON, err := l.taskboard.Pop(l.qName)
			if err != nil {
				l.log.Println("cannot pop job from queue", err.Error())
				continue
			}

			err = json.Unmarshal(jobJSON, &job)
			if err != nil {
				l.log.Println("cannot unmarshal redis output to Job", err.Error())
				continue
			}
			l.log.Printf("unmarshaled to Job %+v", job)
			l.log.Println("sending job to worker channel...")
			l.jobsToRun <- job
			l.log.Println("job sent to worker channel")
		}
	}
}

func (l *listener) stopWorkers() {
	for _, w := range l.workers {
		w.stop()
	}
}

func (l *listener) removeNWorkers(n int) error {
	if n > l.numWorkers {
		return fmt.Errorf("stopNWorkers: given number of workers to shutdown is greater than current worker count")
	}

	for i := 0; i > n; i++ {
		w := l.workers[len(l.workers)-1]
		w.stop()
		l.workers = l.workers[:len(l.workers)-1]
		l.log.Println("removeNWorkers: stop a worker and removed it from the workers slice")
	}

	return nil
}

func (l *listener) addNWorkers(n int) error {
	// TODO: Add global option for maximum allowed worker count (buffered channel and resource limits)
	if n >= l.numWorkers {
		return fmt.Errorf("addNWorkers: given number of workers to add ")
	}
	for i := 0; i < l.numWorkers; i++ {
		w := NewWorker(l.jobsToRun, l.log)
		l.workers = append(l.workers, w)
		go w.work()
	}

	return nil
}
