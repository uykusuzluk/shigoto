package shigoto

import (
	"log"
	"os"
	"sync"
)

type FileTaskboard struct {
	l      *log.Logger
	qfiles map[string]*QueueFile
}

type QueueFile struct {
	mu   sync.Mutex
	file string
}

func (f *FileTaskboard) Initialize(l *log.Logger) error {
	f.qfiles = make(map[string]*QueueFile)
	f.l = l
	return nil
}

func (f *FileTaskboard) Push(job string, queue string) error {
	_, exists := f.qfiles[queue]
	if !exists {
		file, err := os.Create(queue + ".q")
		if err != nil {
			return err
		}
		err = file.Close()
		if err != nil {
			return err
		}
		f.qfiles[queue] = &QueueFile{file: queue}
	}

	qf := f.qfiles[queue]
	qf.mu.Lock()
	defer qf.mu.Unlock()

	file, err := os.OpenFile(queue+".q", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if _, err := file.Write([]byte(job + "\n")); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func (f *FileTaskboard) Pop(queue string) ([]byte, error) {
	// TODO: pop the first line from the file??
	return []byte{}, nil
}

func (f *FileTaskboard) Close() error {
	return nil
}
