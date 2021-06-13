package shigoto

type Runner interface {
	Run() error
}

type Queuer interface {
	Queue() ([]byte, error)
}

type QueueRunner interface {
	Queuer
	Runner
}
