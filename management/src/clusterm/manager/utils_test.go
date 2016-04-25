// +build unittest

package manager

import (
	"github.com/contiv/errored"
	. "gopkg.in/check.v1"
)

type eventUtilsSuite struct {
}

var _ = Suite(&eventUtilsSuite{})

func recordCb(strs *[]string) setInvStateCallback {
	return func(name string) error {
		*strs = append(*strs, name)
		return nil
	}
}

func failureCb(strs *[]string, failOn int) setInvStateCallback {
	return func(name string) error {
		*strs = append(*strs, name)
		if len(*strs) == failOn {
			return errored.Errorf("test failure")
		}
		return nil
	}
}

func (s *eventUtilsSuite) TestSetStatusAtomicSuccess(c *C) {
	strs := []string{"foo", "bar", "dead", "beef"}
	setStrs := []string{}
	revertStrs := []string{}
	mgr := &Manager{}
	mgr.setAssetsStatusAtomic(strs, recordCb(&setStrs), recordCb(&revertStrs))
	c.Assert(strs, DeepEquals, setStrs)
	c.Assert(len(revertStrs), Equals, 0)
}

func (s *eventUtilsSuite) TestSetStatusAtomicFailure(c *C) {
	strs := []string{"foo", "bar", "dead", "beef", "test", "blah"}
	setStrs := []string{}
	revertStrs := []string{}
	mgr := &Manager{}
	mgr.setAssetsStatusAtomic(strs, failureCb(&setStrs, 2), recordCb(&revertStrs))
	c.Assert(len(setStrs), Equals, 2)
	c.Assert(setStrs, DeepEquals, revertStrs)
}

func (s *eventUtilsSuite) TestSetStatusBestEffortSuccess(c *C) {
	strs := []string{"foo", "bar", "dead", "beef"}
	setStrs := []string{}
	mgr := &Manager{}
	mgr.setAssetsStatusBestEffort(strs, recordCb(&setStrs))
	c.Assert(strs, DeepEquals, setStrs)
}

func (s *eventUtilsSuite) TestSetStatusBestEffortFailure(c *C) {
	strs := []string{"foo", "bar", "dead", "beef", "test", "blah"}
	setStrs := []string{}
	mgr := &Manager{}
	mgr.setAssetsStatusBestEffort(strs, failureCb(&setStrs, 2))
	c.Assert(strs, DeepEquals, setStrs)
}
