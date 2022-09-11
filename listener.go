package shigoto

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Listener is the single point of loading the channel with jobs from a queue source,
// to send them to the registered active workers
type listener struct {
	queue string
	state listenerState

	log       *log.Logger
	taskboard TaskBoard

	workers []*worker

	// TODO: turn it into a send-only channel?
	jobsChan  chan Job
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

func newListener(queue string, wcount int, s *Shigoto) (*listener, error) {
	l := &listener{
		log:       s.log,
		taskboard: s.taskBoard,
		queue:     queue,
		state:     initializing,
		pauseChan: make(chan struct{}),
		stopChan:  make(chan struct{}),
	}
	l.init(wcount)
	return l, nil
}

func (l *listener) init(wcount int) {
	jobChan := make(chan Job)
	l.jobsChan = jobChan
	for i := 0; i < wcount; i++ {
		w := newWorker(jobChan, l.log)
		l.workers = append(l.workers, w)
		go w.work()
	}
	l.state = listening
}

func newListenerForWeaver(queue string, jobChan chan Job, s *Shigoto) (*listener, error) {
	return &listener{
		log:       s.log,
		taskboard: s.taskBoard,
		queue:     queue,
		state:     initializing,
		jobsChan:  make(chan Job),
		pauseChan: make(chan struct{}),
		stopChan:  make(chan struct{}),
	}, nil
}

func (l *listener) listen() {
	for {
		select {
		case <-l.stopChan:
			l.state = stopped
			l.stopWorkers()
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
	jobJSON, err := l.taskboard.Pop(l.queue)
	if err != nil {
		l.log.Println("cannot pop job from queue: ", l.queue, " err: ", err.Error())
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

func (l *listener) removeWorkers(n int) error {
	if n > len(l.workers) {
		return fmt.Errorf("stopNWorkers: given number of workers to shutdown is greater than current worker count")
	}

	for i := 0; i > n; i++ {
		w := l.workers[len(l.workers)-1]
		w.stop()
		l.workers = l.workers[:len(l.workers)-1]
		l.log.Println("removeNWorkers: stopped and removed the last worker from the workers' slice")
	}

	return nil
}

func (l *listener) addWorkers(n int) error {
	// TODO: Add global option for maximum allowed worker count (buffered channel and resource limits)
	if (len(l.workers) + n) >= 100 {
		return fmt.Errorf("addNWorkers: requested amount of workers exceed the limit: %d workers", n)
	}

	for i := 0; i < n; i++ {
		w := newWorker(l.jobsChan, l.log)
		l.workers = append(l.workers, w)
		go w.work()
	}

	return nil
}

func (l *listener) setWorkerCount(n int) error {
	if n < 0 {
		return fmt.Errorf("setWorkerCount: given worker count cannot be negative")
	}

	target := n - len(l.workers)
	if target < 0 {
		return l.removeWorkers(-target)
	}

	if target > 0 {
		return l.addWorkers(target)
	}

	return nil
}

func (l *listener) stopWorkers() {
	if l.workers == nil {
		return
	}

	for _, w := range l.workers {
		w.stop()
	}
}

// GracefulClose waits for the Job channel to empty or 3 minutes
// (whichever comes first before sending back a message which
// signals the channel can be closed.
func (l *listener) gracefulClose() bool {
	for {
		select {
		case <-time.After(2 * time.Minute):
			return true
		default:
			if l.state != listening && len(l.jobsChan) == 0 {
				return true
			}
		}
	}
}

func (l *listener) close() {
	l.stop()

	if l.workers != nil {
		l.gracefulClose()
	}

	close(l.jobsChan)
	close(l.pauseChan)
	close(l.stopChan)
}
