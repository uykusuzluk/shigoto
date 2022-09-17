package shigoto

import (
	"encoding/json"
	"log"
)

// Listener is the single point of loading the channel with jobs from a queue source,
// to send them to the registered active workers
type listener struct {
	queue string
	state listenerState

	log       *log.Logger
	taskboard TaskBoard

	jobsChan  chan<- Job
	stopChan  chan struct{}
	pauseChan chan struct{}
}

type listenerState int

// States defined for the status of Listener
const (
	initializing listenerState = iota + 1
	listening
	pausing
	paused
	resuming
	stopping
	stopped
)

func newListener(queue string, s *Shigoto, jobChan chan<- Job) *listener {
	return &listener{
		log:       s.log,
		taskboard: s.taskBoard,
		queue:     queue,
		state:     initializing,
		jobsChan:  jobChan,
		pauseChan: make(chan struct{}),
		stopChan:  make(chan struct{}),
	}
}

func (l *listener) listen() {
	for {
		select {
		case <-l.stopChan:
			l.state = stopped
			return
		case <-l.pauseChan:
			if l.state == pausing {
				l.state = paused
				continue
			}

			if l.state == resuming {
				l.state = listening
				continue
			}
		default:
			if l.state != listening {
				continue
			}

			l.run()
		}
	}
}

func (l *listener) run() {
	job := Job{}
	jobJSON, err := l.taskboard.QRead(l.queue)
	if err != nil {
		l.log.Println("cannot read job from queue: ", l.queue, " err: ", err.Error())
		return
	}

	err = json.Unmarshal(jobJSON, &job)
	if err != nil {
		l.log.Println("cannot unmarshal redis output to Job", err.Error())
		// TODO: Run failed job routines
		return
	}
	//l.log.Printf("unmarshaled to Job %+v", job)
	l.log.Println("listen: sending job to worker channel... q: ", l.queue)
	l.jobsChan <- job
	l.log.Println("listen: job sent to worker channel. q: ", l.queue)
}

func (l *listener) pause() {
	if l.state == listening {
		l.state = pausing
		l.pauseChan <- struct{}{}
	}
}

func (l *listener) unpause() {
	if l.state == paused {
		l.state = resuming
		l.pauseChan <- struct{}{}
	}
}

func (l *listener) stop() {
	l.state = stopping
	l.stopChan <- struct{}{}
}

func (l *listener) close() {
	l.stop()
	close(l.pauseChan)
	close(l.stopChan)
}
