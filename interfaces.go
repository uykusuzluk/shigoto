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

type QNameNewRunner interface {
	QNamer
	Newer
	Runner
}

// JSONer interface enforces a method for the serialization of
// the type to a JSON with its own proper method.
// This way types that require unexported fields for encapsulation
// can keep those fields, and the serialization made by its own
// method instead of json.Marshal() can be used to properly
// move the data to a queue backend.
type JSONer interface {
	JSON() ([]byte, error)
}

// Identifier interface enforces a method for the type information
// of the task. Implementing this method will allow easy serialization
// of the object even by the workers coded in different programming languages.
type Identifier interface {
	Identify() string
}
