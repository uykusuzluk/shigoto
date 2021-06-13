package shigoto

import "log"

type TaskBoard interface {
	Initialize(*log.Logger) error
	Push(string, string) error
	Pop(string) ([]byte, error)
	Close()
}
