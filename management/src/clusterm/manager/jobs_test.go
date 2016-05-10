// +build unittest

package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/contiv/errored"

	. "gopkg.in/check.v1"
)

type jobsSuite struct {
}

var _ = Suite(&jobsSuite{})

func expectDoneCb(c *C, sigCalled chan struct{}, exptdStatus JobStatus, exptdErr error) DoneCallback {
	return func(status JobStatus, jobErr error) {
		c.Assert(status, Equals, exptdStatus)
		c.Assert(jobErr, DeepEquals, exptdErr)
		// signal that the call back was called
		sigCalled <- struct{}{}
	}
}

func runner(wg *sync.WaitGroup, timeout time.Duration, retErr error) JobRunner {
	return func(cancelCh CancelChannel, logs io.Writer) error {
		<-time.After(timeout)
		defer wg.Done()
		return retErr
	}
}

func cancellableRunner(c *C, wg *sync.WaitGroup, timeout time.Duration, retErr error) JobRunner {
	return func(cancelCh CancelChannel, logs io.Writer) error {
		select {
		case <-cancelCh:
			defer wg.Done()
			return retErr
		case <-time.After(timeout):
			c.Assert(false, Equals, true, Commentf("timeout waiting for cancel signal"))
			return nil
		}
	}
}

func logRunner(c *C, wg *sync.WaitGroup, logStr string) JobRunner {
	return func(cancelCh CancelChannel, logs io.Writer) error {
		_, _ = logs.Write([]byte(logStr))
		// checking for error causes test to deadlock. So we just check for logged out in caller
		//c.Assert(err, NotNil)
		defer wg.Done()
		return nil
	}
}

func logRunnerLong(c *C, wg *sync.WaitGroup, timeout time.Duration, logStr1, logStr2 string) JobRunner {
	return func(cancelCh CancelChannel, logs io.Writer) error {
		_, _ = logs.Write([]byte(logStr1))
		<-time.After(timeout)
		_, _ = logs.Write([]byte(logStr2))
		defer wg.Done()
		return nil
	}
}

func waitAndCheckJobStatus(c *C, wg *sync.WaitGroup, job *Job, exptdStatus JobStatus, exptdErr error) {
	wg.Wait()

	// wait for sometime to ensure runner has completly stopped
	<-time.After(100 * time.Millisecond)

	status, errRet := job.Status()
	c.Assert(status, Equals, exptdStatus)
	if exptdErr == nil {
		c.Assert(errRet, IsNil)
	} else {
		c.Assert(errRet, NotNil)
		c.Assert(errRet.Error(), Equals, exptdErr.Error())
	}
}

func checkDoneCb(c *C, cbCh chan struct{}) {
	select {
	case <-cbCh:
	case <-time.After(100 * time.Millisecond):
		c.Assert(false, Equals, true, Commentf("didn't receive job completion callback"))
	}
}

func (s *jobsSuite) TestJobRunSuccess(c *C) {
	wg := &sync.WaitGroup{}
	cbCh := make(chan struct{}, 1)
	j := NewJob(runner(wg, 0, nil), expectDoneCb(c, cbCh, Complete, nil))
	wg.Add(1)
	go j.Run()

	waitAndCheckJobStatus(c, wg, j, Complete, nil)

	checkDoneCb(c, cbCh)
}

func (s *jobsSuite) TestJobStatusRunning(c *C) {
	wg := &sync.WaitGroup{}
	cbCh := make(chan struct{}, 1)
	j := NewJob(runner(wg, 3*time.Second, nil), expectDoneCb(c, cbCh, Complete, nil))
	wg.Add(1)
	go j.Run()
	// give some time for job to start
	time.Sleep(1 * time.Second)
	status, errRet := j.Status()
	c.Assert(status, Equals, Running)
	c.Assert(errRet, Equals, nil)

	waitAndCheckJobStatus(c, wg, j, Complete, nil)

	checkDoneCb(c, cbCh)
}

func (s *jobsSuite) TestJobRunErrored(c *C) {
	wg := &sync.WaitGroup{}
	cbCh := make(chan struct{}, 1)
	err := errored.Errorf("test job failure")
	j := NewJob(runner(wg, 0, err), expectDoneCb(c, cbCh, Errored, err))
	wg.Add(1)
	go j.Run()

	waitAndCheckJobStatus(c, wg, j, Errored, err)

	checkDoneCb(c, cbCh)
}

func (s *jobsSuite) TestJobRunCancel(c *C) {
	wg := &sync.WaitGroup{}
	cbCh := make(chan struct{}, 1)
	err := errored.Errorf("test job cancellation")
	j := NewJob(cancellableRunner(c, wg, 3*time.Second, err), expectDoneCb(c, cbCh, Errored, err))
	wg.Add(1)
	go j.Run()
	// give some time for job to start
	time.Sleep(1 * time.Second)
	c.Assert(j.Cancel(), IsNil)

	waitAndCheckJobStatus(c, wg, j, Errored, err)

	checkDoneCb(c, cbCh)
}

func (s *jobsSuite) TestJobLogs(c *C) {
	wg := &sync.WaitGroup{}
	cbCh := make(chan struct{}, 1)
	exptdLogStr := `
	foo
	bar 1 2 3
	multi line.
	`
	j := NewJob(logRunner(c, wg, exptdLogStr), expectDoneCb(c, cbCh, Complete, nil))
	wg.Add(1)
	go j.Run()

	waitAndCheckJobStatus(c, wg, j, Complete, nil)

	rcvdLogs, err := ioutil.ReadAll(j.Logs())
	c.Assert(err, IsNil)
	c.Assert([]byte(rcvdLogs), DeepEquals, []byte(exptdLogStr))
	// read again to make sure it works everytime
	rcvdLogs, err = ioutil.ReadAll(j.Logs())
	c.Assert(err, IsNil)
	c.Assert([]byte(rcvdLogs), DeepEquals, []byte(exptdLogStr))

	checkDoneCb(c, cbCh)
}

func (s *jobsSuite) TestJobLogsLongRunning(c *C) {
	wg := &sync.WaitGroup{}
	cbCh := make(chan struct{}, 1)
	exptdLogStr1 := `
	foo
	bar 1 2 3
	multi line.
	`
	exptdLogStr2 := `
	foo1
	bar1 1 2 3
	multi line. 1
	`
	j := NewJob(logRunnerLong(c, wg, 3*time.Second, exptdLogStr1, exptdLogStr2), expectDoneCb(c, cbCh, Complete, nil))
	wg.Add(1)
	go j.Run()
	// give some time for job to start and fetch logs once
	time.Sleep(1 * time.Second)
	j.Logs()
	_, _ = ioutil.ReadAll(j.Logs())

	waitAndCheckJobStatus(c, wg, j, Complete, nil)

	rcvdLogs, err := ioutil.ReadAll(j.Logs())
	c.Assert(err, IsNil)
	c.Assert([]byte(rcvdLogs), DeepEquals, []byte(exptdLogStr1+exptdLogStr2))

	checkDoneCb(c, cbCh)
}

func (s *jobsSuite) TestJobInfoMarshal(c *C) {
	exptdLogStr := `
	foo
	bar 1 2 3
	multi line.
	`
	j := &Job{
		status: Running,
		errVal: nil,
		logs:   *bytes.NewBuffer([]byte(exptdLogStr)),
	}

	out, err := j.MarshalJSON()
	c.Assert(err, IsNil)

	// verify the relevant fields
	exptdInfo := struct {
		Status string   `json:"status"`
		ErrVal string   `json:"error"`
		Logs   []string `json:"logs"`
	}{}
	err = json.Unmarshal(out, &exptdInfo)
	c.Assert(err, IsNil)
	c.Assert(exptdInfo.Status, Equals, Running.String())
	c.Assert(exptdInfo.ErrVal, Equals, fmt.Sprintf("%v", j.errVal))
	c.Assert(exptdInfo.Logs, DeepEquals, strings.Split(exptdLogStr, "\n"))
}
