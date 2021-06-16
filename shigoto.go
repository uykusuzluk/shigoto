package shigoto

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
)

var jobContainer map[string]Runner

type Shigoto struct {
	log       *log.Logger
	taskBoard TaskBoard
	listeners []listener
}

type ShigotoOpts func(*Shigoto) error

func New(opts ...ShigotoOpts) *Shigoto {
	s := &Shigoto{}
	s.defaultLogger()

	for _, optFunc := range opts {
		err := optFunc(s)
		if err != nil {
			panic(fmt.Sprintln("Problem with option ", optFunc, " err: ", err.Error()))
		}
	}
	s.initialize()
	return s
}

func (s *Shigoto) initialize() error {
	jobContainer = make(map[string]Runner)

	if s.taskBoard == nil {
		err := s.defaultTaskboard()
		if err != nil {
			return err
		}
	}
	return nil
}

// WithRedis sets the queue connection for Shigoto as Redis backend
func WithRedis() func(*Shigoto) error {
	return func(s *Shigoto) error {
		s.taskBoard = &Redis{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			Database: 0,
		}
		return s.taskBoard.Initialize(s.log)
	}
}

// WithTaskboard sets the taskboard with given backend compliant to the Taskboard interface
func WithTaskboard(t TaskBoard) func(*Shigoto) error {
	return func(s *Shigoto) error {
		s.taskBoard = t
		return s.taskBoard.Initialize(s.log)
	}
}

// WithLogger sets the logger for Shigoto to the given *log.Logger
func WithLogger(l *log.Logger) func(*Shigoto) error {
	return func(s *Shigoto) error {
		s.log = l
		return nil
	}
}

// Close closes necessary used resources to prevent memory leaks
func (s *Shigoto) Close() error {
	s.stopListeners()
	s.taskBoard.Close()
	return nil
}

func (s *Shigoto) stopListeners() {
	for _, l := range s.listeners {
		l.stopChan <- struct{}{}
	}
}

func (s *Shigoto) defaultLogger() {
	s.log = log.New(os.Stdout, "shigoto ", log.LstdFlags)
	s.log.Println("default logger set for shigoto!")
}

func (s *Shigoto) defaultTaskboard() error {
	s.taskBoard = &Redis{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		Database: 0,
	}
	return s.taskBoard.Initialize(s.log)
}

// QueueDefault makes it possible to queue a job with implied default queue name from a QName() method
func (s *Shigoto) QueueDefault(job QNameRunner) error {
	return s.Queue(job, job.QName())
}

func (s *Shigoto) Queue(job Runner, queue string) error {
	encodedJob, err := json.Marshal(job)
	if err != nil {
		return err
	}

	jobForQ, err := NewJob(encodedJob, reflect.TypeOf(job).String(), queue)
	if err != nil {
		return err
	}
	s.taskBoard.Push(jobForQ, queue)

	return nil
}

func (s *Shigoto) ListenQueue(queue string, workers int) error {
	listener, err := newListener(queue, workers, s)
	if err != nil {
		return err
	}
	s.listeners = append(s.listeners, *listener)
	go listener.Listen()
	return nil
}

func (s *Shigoto) Register(j Runner) error {
	s.log.Println("register: type of runner is: ", reflect.TypeOf(j).String())
	jobContainer[reflect.TypeOf(j).String()] = j
	return nil
}
