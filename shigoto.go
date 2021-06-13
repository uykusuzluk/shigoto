package shigoto

import (
	"encoding/json"
	"log"
	"os"
	"reflect"

	"github.com/go-redis/redis"
)

var jobContainer map[string]Runner

type Shigoto struct {
	log       *log.Logger
	redis     *redis.Client
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

	err := s.setRedis()
	if err != nil {
		return err
	}

	return nil
}

func (s *Shigoto) Close() error {
	s.stopListeners()
	s.redis.Close()
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

func (s *Shigoto) setRedis() error {
	s.log.Println("connecting to redis...")
	s.redis = redis.NewClient(&redis.Options{
		Addr:     "localhost" + ":" + "6379",
		Password: "",
		DB:       0,
	})

	redisStatus := s.redis.Ping()
	if err := redisStatus.Err(); err != nil {
		s.log.Fatalln("cannot ping Redis: ", err.Error())
		return err
	}
	s.log.Println("connected to redis!")
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
	s.redis.RPush(queue, jobForQ)

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
	jobContainer[reflect.TypeOf(j).String()] = j
	return nil
}
