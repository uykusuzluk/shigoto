package shigoto

// Runner interface defines runnable jobs with Run() methods
type Runner interface {
	Run() error
}

type Newer interface {
	New() (Runner, error)
}

type NewRunner interface {
	Newer
	Runner
}

// QNamer interface allows implied default queue name for jobs
type QNamer interface {
	QName() string
}

// QNameRunner interface that combines Runner and QNamer interfaces
type QNameRunner interface {
	QNamer
	Runner
}

type Tasker interface {
	QNamer
	Newer
	Runner
}

// Identifier interface enforces a method for the type information
// of the task. Implementing this method can allow easy serialization
// of the object even by the workers coded in different programming languages.
type Identifier interface {
	Identify() string
}
