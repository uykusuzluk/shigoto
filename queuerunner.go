package shigoto

type Runner interface {
	Run() error
}

type QNamer interface {
	QName() string
}

type Queuer interface {
	Queue() ([]byte, error)
}

type QueueRunner interface {
	Queuer
	Runner
}

type QNameRunner interface {
	QNamer
	Runner
}
