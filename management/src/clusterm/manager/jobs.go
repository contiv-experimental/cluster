package manager

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"

	"github.com/contiv/errored"
)

// CancelChannel is type of the channle used to signal cancellation of job
type CancelChannel chan struct{}

// JobRunner is blocking call that runs the task. On receiving a signal on the
// cancel-channel, it is expected to return immediately
type JobRunner func(cancelCh CancelChannel) error

// DoneCallback is called when job completes, errors or is cancelled
type DoneCallback func(status JobStatus, errRet error)

// Job corresponds to a long running task, triggered by an event
// XXX: add log buffer support to store job specific logs
type Job struct {
	sync.Mutex
	runner   JobRunner
	done     DoneCallback
	cancelCh CancelChannel
	status   JobStatus
	errRet   error
}

// NewJob initializes and returns an instance of a job decribed by the runner and done callback
func NewJob(jr JobRunner, done DoneCallback) *Job {
	return &Job{
		runner:   jr,
		done:     done,
		cancelCh: make(chan struct{}),
		status:   Queued,
		errRet:   nil,
	}
}

// String returns a brief description of the job
func (j *Job) String() string {
	runnerName := runtime.FuncForPC(reflect.ValueOf(j.runner).Pointer()).Name()
	return fmt.Sprintf("[task: %s status: %v errRet: %v]", runnerName, j.status, j.errRet)
}

func (j *Job) setStatus(status JobStatus, err error) {
	j.Lock()
	j.status = status
	j.errRet = err
	j.Unlock()
}

// Run begins the job and wait for completion. This function blocks
func (j *Job) Run() {
	var err error

	j.setStatus(Running, nil)
	defer func() {
		j.done(j.status, j.errRet)
	}()

	if err = j.runner(j.cancelCh); err != nil {
		j.setStatus(Errored, err)
		return
	}
	j.setStatus(Complete, nil)
}

//Cancel signals canceling a running job
func (j *Job) Cancel() error {
	// if job is running then run it's cancel function
	// the job status shall be updated as part of runner
	j.Lock()
	defer j.Unlock()
	if j.status == Running {
		j.cancelCh <- struct{}{}
		return nil
	}
	return errored.Errorf("job is not Running")
}

// Status returns the status of a job at the time of call
func (j *Job) Status() (JobStatus, error) {
	return j.status, j.errRet
}
