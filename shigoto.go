package shigoto

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
)

var jobContainer map[string]NewRunner

type Shigoto struct {
	log       *log.Logger
	taskBoard TaskBoard
	listeners []listener
}

// ShigotoOpts is the option function type for setting the properties of a Shigoto instance
type ShigotoOpts func(*Shigoto) error

// New creates a new Shigoto instance for queueing and processing jobs
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
	jobContainer = make(map[string]NewRunner)

	if s.taskBoard == nil {
		err := s.defaultTaskboard()
		if err != nil {
			return err
		}
	}
	return nil
}

// WithRedis sets the queue connection for Shigoto as Redis backend
func WithRedis() ShigotoOpts {
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
func WithTaskboard(t TaskBoard) ShigotoOpts {
	return func(s *Shigoto) error {
		s.taskBoard = t
		return s.taskBoard.Initialize(s.log)
	}
}

// WithLogger sets the logger for Shigoto to the given *log.Logger
func WithLogger(l *log.Logger) ShigotoOpts {
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
	redisOpt := WithRedis()
	return redisOpt(s)
}

// Queue makes it possible to queue a job with implied default queue name from a QName() method
func (s *Shigoto) Queue(job QNameNewRunner) error {
	return s.QueueTo(job, job.QName())
}

func (s *Shigoto) QueueTo(job NewRunner, queue string) error {
	var (
		payload []byte
		err     error
	)

	if pl, isJSONer := job.(JSONer); isJSONer {
		payload, err = pl.JSON()
	} else {
		payload, err = json.Marshal(job)
	}

	if err != nil {
		return err
	}

	jobForQ, err := newJob(payload, reflect.TypeOf(job).String(), queue)
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
	go listener.listen()
	return nil
}

func (s *Shigoto) Register(j NewRunner) error {
	if jobIDer, isIdentifier := j.(Identifier); isIdentifier {
		jobContainer[jobIDer.Identify()] = j
		s.log.Println("register: type of runner is: ", jobIDer.Identify())
		return nil
	}

	name := reflect.TypeOf(j).String()
	s.log.Println("register: type of runner is: ", name)
	jobContainer[name] = j
	return nil
}
