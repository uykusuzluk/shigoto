package shigoto

// Runner interface defines runnable jobs with Run() methods
type Runner interface {
	Run() error
}

// QNamer interface allows implied default queue name for jobs
type QNamer interface {
	QName() string
}

// type Queuer interface {
// 	Queue() ([]byte, error)
// }

// type QueueRunner interface {
// 	Queuer
// 	Runner
// }

// QNameRunner interface that combines Runner and QNamer interfaces
type QNameRunner interface {
	QNamer
	Runner
}
