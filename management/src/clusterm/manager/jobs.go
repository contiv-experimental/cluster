package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/contiv/errored"
)

var notRunningErr = errored.Errorf("job is not Running")

// CancelChannel is type of the channle used to signal cancellation of job
type CancelChannel chan struct{}

// JobRunner is blocking call that runs the task. On receiving a signal on the
// cancel-channel, it is expected to return immediately
type JobRunner func(cancelCh CancelChannel, logs io.Writer) error

// DoneCallback is called when job completes, errors or is cancelled
type DoneCallback func(status JobStatus, errVal error)

// Job corresponds to a long running task, triggered by an event
type Job struct {
	sync.Mutex
	runner    JobRunner
	done      DoneCallback
	cancelCh  CancelChannel
	status    JobStatus
	errVal    error
	logs      bytes.Buffer
	logWriter *MultiWriter
	desc      string
}

// NewJob initializes and returns an instance of a job described by the runner and done callback
func NewJob(desc string, jr JobRunner, done DoneCallback) *Job {
	j := &Job{
		runner:    jr,
		done:      done,
		desc:      desc,
		cancelCh:  make(chan struct{}),
		status:    Queued,
		errVal:    nil,
		logWriter: &MultiWriter{},
	}
	j.logWriter.Add(&j.logs)
	return j
}

func (j *Job) runnerName() string {
	return runtime.FuncForPC(reflect.ValueOf(j.runner).Pointer()).Name()
}

// String returns a brief description of the job
func (j *Job) String() string {
	return fmt.Sprintf("[task: %s status: %v errVal: %v]", j.runnerName(), j.status, j.errVal)
}

func (j *Job) setStatus(status JobStatus, err error) {
	j.Lock()
	j.status = status
	j.errVal = err
	j.Unlock()
}

// Run begins the job and wait for completion. This function blocks
func (j *Job) Run() {
	j.setStatus(Running, nil)
	defer func() {
		j.done(j.status, j.errVal)
		j.logWriter.Close()
	}()

	if err := j.runner(j.cancelCh, j.logWriter); err != nil {
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
	return notRunningErr
}

// Status returns the status of a job at the time of call
func (j *Job) Status() (JobStatus, error) {
	return j.status, j.errVal
}

// Logs returns the current logs associated with the job.
func (j *Job) Logs() io.Reader {
	// instead of returning the buffer itself we instead need to return
	// a reader created over current contents of the buffer without changing
	// it's read offset. This will allow accessing logs over and over again.
	return bytes.NewReader(j.logs.Bytes())
}

// PipeLogs pipes the job logs to the specified writer (in addition to underlying log buffer).
// This is useful to stream ongoing job logs to additional writer(s).
func (j *Job) PipeLogs(w io.Writer) error {
	if s, _ := j.Status(); s != Running {
		return notRunningErr
	}
	j.logWriter.Add(w)
	return nil
}

// MarshalJSON marshals and returns the JSON for job info
func (j *Job) MarshalJSON() ([]byte, error) {
	toJSON := struct {
		Desc   string   `json:"desc"`
		Task   string   `json:"task"`
		Status string   `json:"status"`
		ErrVal string   `json:"error"`
		Logs   []string `json:"logs"`
	}{
		Desc:   j.desc,
		Task:   j.runnerName(),
		Status: j.status.String(),
		Logs:   strings.Split(j.logs.String(), "\n"),
	}
	if j.errVal != nil {
		toJSON.ErrVal = fmt.Sprintf("%v", j.errVal)
	}

	return json.Marshal(toJSON)
}
