package shigoto

import (
	"encoding/json"
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

	jobsToRun chan<- Job
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

func (l *listener) Listen() {
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
