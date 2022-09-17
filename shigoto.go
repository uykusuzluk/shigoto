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

func (s *Shigoto) defaultLogger() {
	s.log = log.New(os.Stdout, "shigoto ", log.LstdFlags)
	s.log.Println("default logger set for shigoto!")
}

func (s *Shigoto) defaultTaskboard() error {
	redisOpt := WithRedis()
	return redisOpt(s)
}

func (s *Shigoto) stopListeners() {
	for _, l := range s.listeners {
		l.close()
	}
}

// Queue makes it possible to queue a job with implied default queue name from a QName() method
func (s *Shigoto) Queue(task Tasker) error {
	return s.QueueTo(task, task.QName())
}

// QueueTo allows the queueing of a job to the desired queue name
func (s *Shigoto) QueueTo(job NewRunner, queue string) error {
	var (
		payload []byte
		err     error
		jobForQ string
	)

	if pl, isMarshaler := job.(json.Marshaler); isMarshaler {
		payload, err = pl.MarshalJSON()
	} else {
		payload, err = json.Marshal(job)
	}

	if err != nil {
		return err
	}

	if ident, isIdentifier := job.(Identifier); isIdentifier {
		jobForQ, err = newJob(payload, ident.Identify(), queue)
	} else {
		jobForQ, err = newJob(payload, reflect.TypeOf(job).String(), queue)
	}

	if err != nil {
		return err
	}
	s.taskBoard.QWrite(jobForQ, queue)

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

// Register adds a NewRunner type to the job container so that it can be processed by workers
func (s *Shigoto) Register(j NewRunner) error {
	if jobIDer, isIdentifier := j.(Identifier); isIdentifier {
		s.registerAs(j, jobIDer.Identify())
		return nil
	}

	s.registerAs(j, reflect.TypeOf(j).String())
	return nil
}

// RegisterAs adds a NewRunner type to the job container so that it can be processed by workers
func (s *Shigoto) registerAs(j NewRunner, payloadType string) {
	s.log.Println("register: type of runner is: ", payloadType)
	jobContainer[payloadType] = j
}

// Close used resources to prevent memory leaks
func (s *Shigoto) Close() error {
	s.stopListeners()
	s.taskBoard.Close()
	return nil
}
