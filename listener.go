package shigoto

import (
	"encoding/json"
	"fmt"
	"log"
)

// Listener is the single point of loading the channel with jobs from a queue source,
// to send them to the registered active workers
type listener struct {
	qName string
	state listenerState

	log       *log.Logger
	taskboard TaskBoard

	workers []*Worker

	// TODO: turn it into a send-only channel?
	jobsToRun chan Job
	stopChan  chan struct{}
	pauseChan chan struct{}
}

type listenerState int

// States defined for the status of Listener
const (
	configuring listenerState = iota + 1
	running
	pausing
	resuming
	paused
	closing
)

func newListener(queue string, wcount int, s *Shigoto) (*listener, error) {
	l := &listener{
		log:       s.log,
		taskboard: s.taskBoard,
		qName:     queue,
		state:     configuring,
	}
	l.init(wcount)
	return l, nil
}

func (l *listener) init(wcount int) {
	jobChan := make(chan Job, wcount*2)
	l.jobsToRun = jobChan
	for i := 0; i < wcount; i++ {
		w := newWorker(jobChan, l.log)
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
		case <-l.pauseChan:
			if l.state == pausing {
				l.state = paused
				continue
			}

			if l.state == resuming {
				l.state = running
				continue
			}
		default:
			if l.state != running {
				continue
			}

			l.run()
		}
	}
}

func (l *listener) run() {
	job := Job{}
	jobJSON, err := l.taskboard.Pop(l.qName)
	if err != nil {
		l.log.Println("cannot pop job from queue: ", l.qName, " err: ", err.Error())
		return
	}

	err = json.Unmarshal(jobJSON, &job)
	if err != nil {
		l.log.Println("cannot unmarshal redis output to Job", err.Error())
		// TODO: Run failed job routines
		return
	}
	//l.log.Printf("unmarshaled to Job %+v", job)
	l.log.Println("listen: sending job to worker channel... q: ", l.qName)
	l.jobsToRun <- job
	l.log.Println("listen: job sent to worker channel. q: ", l.qName)
}

func (l *listener) pause() {
	if l.state == running {
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
	l.state = closing
	l.stopChan <- struct{}{}
}

func (l *listener) stopWorkers() {
	for _, w := range l.workers {
		w.stop()
	}
}

func (l *listener) removeNWorkers(n int) error {
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

func (l *listener) addNWorkers(n int) error {
	// TODO: Add global option for maximum allowed worker count (buffered channel and resource limits)
	if (len(l.workers) + n) >= 100 {
		return fmt.Errorf("addNWorkers: requested amount of workers exceed the limit: %d workers", n)
	}

	for i := 0; i < n; i++ {
		w := newWorker(l.jobsToRun, l.log)
		l.workers = append(l.workers, w)
		go w.work()
	}

	return nil
}

func (l *listener) modWorkerCount(n int) error {
	if n < 0 {
		return fmt.Errorf("modWorkerCount: given worker count cannot be negative")
	}

	target := n - len(l.workers)
	if target < 0 {
		return l.removeNWorkers(-target)
	}

	if target > 0 {
		return l.addNWorkers(target)
	}

	return nil
}
