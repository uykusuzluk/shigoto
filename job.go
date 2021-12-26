package shigoto

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"time"

	"github.com/google/uuid"
)

const (
	datetimeFormat = "2006-01-02 15:04:05"
)

// Job is the main struct that holds the payload to be queued
type Job struct {
	UUID  string
	QName string

	Status  JobStatus
	Attemps int

	MaxAttemps  int
	MaxRunTime  time.Duration
	ExpireAfter time.Duration
	// Delay       time.Duration

	QueuedAt   time.Time
	StartedAt  time.Time
	FailedAt   time.Time
	FinishedAt time.Time

	PayloadType string
	Payload     []byte
	ErrCheck    uint32
}

// JobStatus is the enumeration for the current status of the job
type JobStatus int

// Enumeration for the JobStatus
const (
	Queued = iota + 1
	Running
	Canceled
	Failed
	Finished
)

func newJob(payload []byte, ptype string, queue string) (string, error) {
	job := Job{
		UUID:        uuid.New().String(),
		QName:       queue,
		Status:      Queued,
		QueuedAt:    time.Now().UTC(),
		PayloadType: ptype,
		Payload:     payload,
		ErrCheck:    crc32.ChecksumIEEE(payload),
	}
	fmt.Println(time.Now().UTC())
	jobJSON, err := json.Marshal(job)
	if err != nil {
		return "", err
	}
	return string(jobJSON), nil
}

// Expired returns true if the current time is later than the sum of job's queued time and expireAfter duration
func (j *Job) expired() (bool, error) {
	if j.ExpireAfter == 0 {
		return false, nil
	}
	if time.Now().After(j.QueuedAt.Local().Add(j.ExpireAfter)) {
		return true, j.failExpired()
	}
	return false, nil
}

// FailExpired method is called when the job has expired.
func (j *Job) failExpired() error {
	return j.failWithMsg("job has expired. queue time: " +
		j.QueuedAt.Local().Format(datetimeFormat) +
		" expireAfter time: " + j.QueuedAt.Local().Add(j.ExpireAfter).Format(datetimeFormat) +
		" time now: " + time.Now().Format(datetimeFormat))
}

func (j *Job) failChecksum() error {
	return j.failWithMsg("checksum hash for payload does not match the current one")
}

func (j *Job) failWithMsg(msg string) error {
	// TODO: write the failure to somewhere
	return fmt.Errorf(msg)
}

func (j *Job) checkPayload() (bool, error) {
	if j.ErrCheck == 0 {
		return true, nil
	}
	if j.ErrCheck == crc32.ChecksumIEEE([]byte(j.Payload)) {
		return true, nil
	}

	return false, j.failChecksum()
}
