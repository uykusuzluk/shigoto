package shigoto

import (
	"encoding/json"
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

func New() *Shigoto {
	s := Shigoto{}
	s.initialize()
	return &s
}

func (s *Shigoto) initialize() error {
	s.setLogger()
	jobContainer = make(map[string]Runner)

	err := s.setTaskboard()
	if err != nil {
		return err
	}

	return nil
}

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

func (s *Shigoto) setLogger() {
	s.log = log.New(os.Stdout, "shigoto ", log.LstdFlags)
	s.log.Println("logger set for shigoto!")
}

func (s *Shigoto) setTaskboard() error {
	s.taskBoard = &Redis{
		Host:     "localhost",
		Port:     "6379",
		Password: "",
		Database: 0,
	}
	s.taskBoard.Initialize(s.log)
	return nil
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
	listener, err := newListener(queue, 10, s)
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
