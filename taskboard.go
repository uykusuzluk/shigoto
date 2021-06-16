package shigoto

import "log"

// TaskBoard interface defines the blueprint for backends to be used
type TaskBoard interface {
	Initialize(*log.Logger) error
	Push(string, string) error
	Pop(string) ([]byte, error)
	Close()
}
