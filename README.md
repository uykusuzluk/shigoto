# shigoto
A distributed job/task scheduling/queueing library

## Basic Usage
```golang
package main

// TestJob has only "exported" fields which can be encoded to JSON
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

    // Queue a job
    job1 := TestJob{
        Name: "test job",
        Value: 12345,
    }
    manager.QueueTo(&job1, "testjob_queue")

    // Register a struct as a runnable Job
    manager.Register(&TestJob{})

    // Run job workers by creating Listeners
    manager.ListenQueue("testjob_queue", 2)
}
```
### TODO:  
Monitoring:  
    Node:  
    - Show each listen (with qname) and their worker counts  
    - Allow adding/removing workers or set a number for worker counts  
    - Connect to a master node??? (Need ID first to identify)  
          
Master:  
    - Show each node  
    - Show aggregated workers for a queue  
      
      
**Structure Fields should be exported (JSON Marshaling)  
**jobContainer global  
**remove reflection - make method for naming job "func (a *A) JobName() string"  