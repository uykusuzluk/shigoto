package shigoto

import (
	"log"

	"github.com/go-redis/redis"
)

type Redis struct {
	Host     string
	Port     string
	Password string
	Database int

	redis *redis.Client
	log   *log.Logger
}

func (r *Redis) Initialize(l *log.Logger) error {
	r.log = l
	return r.connect()
}

func (r *Redis) connect() error {
	r.log.Println("connecting to redis...")
	r.redis = redis.NewClient(&redis.Options{
		Addr:     r.Host + ":" + r.Port,
		Password: r.Password,
		DB:       r.Database,
	})

	redisStatus := r.redis.Ping()
	if err := redisStatus.Err(); err != nil {
		r.log.Fatalln("cannot ping Redis: ", err.Error())
		return err
	}
	r.log.Println("connected to redis!")
	return nil
}
func (r *Redis) Push(job string, queue string) error {
	cmd := r.redis.RPush(queue, job)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

func (r *Redis) Pop(queue string) ([]byte, error) {
	listElement := r.redis.BLPop(0, queue)
	if listElement.Err() != nil {
		return nil, listElement.Err()
	}
	return []byte(listElement.Val()[1]), nil
}

func (r *Redis) Close() {
	r.redis.Close()
}
