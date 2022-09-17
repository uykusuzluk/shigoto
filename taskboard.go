package shigoto

import "log"

// TaskBoard interface defines the blueprint for backends to be used
type TaskBoard interface {
	Initialize(*log.Logger) error
	QWriter
	QReader
	Close() error
}

type QWriter interface {
	QWrite(string, string) error
}

type QReader interface {
	QRead(string) ([]byte, error)
}
