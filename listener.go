package shigoto

import (
	"encoding/json"
	"fmt"
	"log"
)

// Listener is the single point of loading the channel with jobs from a queue source,
// to send them to the registered active workers
type listener struct {
	queue string
	state listenerState

	log       *log.Logger
	taskboard TaskBoard

	wp *workerPool
	// TODO: turn it into a send-only channel?
	stopChan  chan struct{}
	pauseChan chan struct{}
}

type listenerState int

// States defined for the status of Listener
const (
	initializing listenerState = iota + 1
	running
	pausing
	resuming
	paused
	closing
	stopped
)

func newListener(queue string, wcount int, s *Shigoto) (*listener, error) {
	l := &listener{
		log:       s.log,
		taskboard: s.taskBoard,
		queue:     queue,
		state:     initializing,
	}
	l.init(wcount)
	return l, nil
}

func (l *listener) init(wcount int) {
	l.wp = newWorkerPool(wcount)
	l.state = running
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
	worker := l.wp.get()
	go worker(job, l.log, l.wp)
	l.log.Println("listen: job sent to worker channel. q: ", l.queue)
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

func (l *listener) close() {
	l.stop()
	for l.state == closing {
		if l.state == stopped {
			break
		}
	}

	l.wp.close()
	close(l.pauseChan)
	close(l.stopChan)
}

func (l *listener) setWorkerCount(n int) error {
	if n < 0 {
		return fmt.Errorf("setWorkerCount: given worker count cannot be negative")
	}

	if n == l.wp.count {
		return nil
	}

	// TODO: below
	l.pause()
	for l.state == pausing {
		if l.state == paused {
			break
		}
	}

	l.wp.close()
	l.wp = newWorkerPool(n)

	return nil
}
