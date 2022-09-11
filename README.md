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

func (t *TestJob) New() (shigoto.Runner, error) {
    return new(TestJob) // &TestJob{}
}

// jobmanager.Register(job)
//           .RegisterAs(job, payloadType)
//           .Queue(job)
//           .QueueTo(job, qName)

func main() {
    // Initialize Shigoto Job Manager
    jm := shigoto.New()

    // Queue a job
    j := TestJob{
        Name: "test job",
        Value: 12345,
    }
    jm.QueueTo(&job1, "testjob_queue")

    // Register a struct as a runnable Job
    jm.Register(&TestJob{})

    // Run job workers by creating Listeners
    jm.ListenQueue("testjob_queue", 2)
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