package shigoto

import (
	"fmt"
	"log"
	"time"
)

type manager struct {
	s        *Shigoto
	queue    string
	jobsChan chan Job
	listener *listener
	workers  []*worker
	log      *log.Logger
}

func newManager(queue string, wcount int, s *Shigoto) *manager {
	jch := make(chan Job)
	m := &manager{
		log:      s.log,
		jobsChan: jch,
		queue:    queue,
		listener: newListener(queue, s, jch),
	}
	m.init(wcount)

	return m
}

func (m *manager) init(wcount int) {
	for i := 0; i < wcount; i++ {
		w := newWorker(m.jobsChan, m.s.log)
		m.workers = append(m.workers, w)
		go w.work()
	}
	m.listener.state = listening
	go m.listener.listen()
}

func (m *manager) close() {
	m.listener.close()
	close(m.jobsChan)

	for _, w := range m.workers {
		w.stop()
		w.close()
	}
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

// GracefulClose waits for the Job channel to empty or 3 minutes
// (whichever comes first before sending back a message which
// signals the channel can be closed.
func (m *manager) gracefulClose() bool {
	for {
		select {
		case <-time.After(2 * time.Minute):
			return true
		default:
			if m.listener.state != listening && len(m.jobsChan) == 0 {
				return true
			}
		}
	}
}
