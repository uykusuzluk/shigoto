# shigoto
A distributed job/task scheduling/queueing library

## Basic Usage
```golang
package main

type TestJob struct {
    Name string
    Value int
}

func (t *TestJob) Run() {
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