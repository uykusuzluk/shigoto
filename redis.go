package shigoto

import (
	"log"

	"github.com/go-redis/redis"
)

// Redis struct to be used as a backend taskboard for queues
type Redis struct {
	Host     string
	Port     string
	Password string
	Database int

	redis *redis.Client
	log   *log.Logger
}

// Initialize prepares Redis backend for use
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

// Push method adds a job to the right end (tail) of the named redis list (queue)
func (r *Redis) QWrite(job string, queue string) error {
	cmd := r.redis.RPush(queue, job)
	if cmd.Err() != nil {
		return cmd.Err()
	}
	return nil
}

// Pop method remove a job from the left end (head) of the named redis list (queue) for processing.
// For redis a blocking function is used on a single listening edge point for a queue name.
func (r *Redis) QRead(queue string) ([]byte, error) {
	listElement := r.redis.BLPop(0, queue)
	if listElement.Err() != nil {
		return nil, listElement.Err()
	}
	return []byte(listElement.Val()[1]), nil
}

// Close prepares the resource for termination without memory leaks
func (r *Redis) Close() error {
	return r.redis.Close()
}
