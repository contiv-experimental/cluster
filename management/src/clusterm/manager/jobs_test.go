// +build unittest

package manager

import (
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
		// signal that the call back was called
		sigCalled <- struct{}{}
		c.Assert(status, Equals, exptdStatus)
		c.Assert(jobErr, DeepEquals, exptdErr)
	}
}

func runner(wg *sync.WaitGroup, timeout time.Duration, retErr error) JobRunner {
	return func(cancelCh CancelChannel) error {
		<-time.After(timeout)
		wg.Done()
		return retErr
	}
}

func cancellableRunner(c *C, wg *sync.WaitGroup, timeout time.Duration, retErr error) JobRunner {
	return func(cancelCh CancelChannel) error {
		select {
		case <-cancelCh:
			wg.Done()
			return retErr
		case <-time.After(timeout):
			c.Assert(false, Equals, true, Commentf("timeout waiting for cancel signal"))
			return nil
		}
	}
}

func (s *jobsSuite) TestJobRunSuccess(c *C) {
	wg := &sync.WaitGroup{}
	cbCh := make(chan struct{}, 1)
	j := NewJob(runner(wg, 0, nil), expectDoneCb(c, cbCh, Complete, nil))
	wg.Add(1)
	go j.Run()
	wg.Wait()
	status, errRet := j.Status()
	c.Assert(status, Equals, Complete)
	c.Assert(errRet, Equals, nil)
	select {
	case <-cbCh:
	case <-time.After(0):
		c.Assert(false, Equals, true, Commentf("didn't receive job completion callback"))
	}
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
	wg.Wait()
	status, errRet = j.Status()
	c.Assert(status, Equals, Complete)
	c.Assert(errRet, Equals, nil)
	select {
	case <-cbCh:
	case <-time.After(0):
		c.Assert(false, Equals, true, Commentf("didn't receive job completion callback"))
	}
}

func (s *jobsSuite) TestJobRunErrored(c *C) {
	wg := &sync.WaitGroup{}
	cbCh := make(chan struct{}, 1)
	err := errored.Errorf("test job failure")
	j := NewJob(runner(wg, 0, err), expectDoneCb(c, cbCh, Errored, err))
	wg.Add(1)
	go j.Run()
	wg.Wait()
	status, errRet := j.Status()
	c.Assert(status, Equals, Errored)
	c.Assert(errRet, DeepEquals, err)
	select {
	case <-cbCh:
	case <-time.After(0):
		c.Assert(false, Equals, true, Commentf("didn't receive job completion callback"))
	}
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
	wg.Wait()
	status, errRet := j.Status()
	c.Assert(status, Equals, Errored)
	c.Assert(errRet, DeepEquals, err)
	select {
	case <-cbCh:
	case <-time.After(0):
		c.Assert(false, Equals, true, Commentf("didn't receive job completion callback"))
	}
}
