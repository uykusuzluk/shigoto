# shigoto
A distributed job/task scheduling/queueing library

## Basic Usage
```golang
package main

type TestJob struct {
    Name string
    Value int
}

func (t *TestJob) Run() error {
    // ...
}

func main() {
    // Initialize Shigoto Job Manager
    manager := shigoto.New()

    // Register a struct as a runnable Job
    manager.Register(&TestJob{})

    // Queue a job
    job1 := TestJob{
        Name: "test job",
        Value: 12345,
    }
    manager.Queue(&job1, "testjob_queue")

    // Run job workers by creating Listeners
    manager.ListenQueue("testjob_queue", 2)
}
```

**Structure Fields should be exported (JSON Marshaling)
**jobContainer global
**remove reflection - make method for naming job "func (a *A) JobName() string"