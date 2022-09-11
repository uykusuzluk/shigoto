package shigoto

import "log"

// WithRedis sets the queue connection for Shigoto as Redis backend
func WithRedis() ShigotoOpts {
	return func(s *Shigoto) error {
		s.taskBoard = &Redis{
			Host:     "localhost",
			Port:     "6379",
			Password: "",
			Database: 0,
		}
		return s.taskBoard.Initialize(s.log)
	}
}

// WithTaskboard sets the taskboard with given backend compliant to the Taskboard interface
func WithTaskboard(t TaskBoard) ShigotoOpts {
	return func(s *Shigoto) error {
		s.taskBoard = t
		return s.taskBoard.Initialize(s.log)
	}
}

// WithLogger sets the logger for Shigoto to the given *log.Logger
func WithLogger(l *log.Logger) ShigotoOpts {
	return func(s *Shigoto) error {
		s.log = l
		return nil
	}
}
