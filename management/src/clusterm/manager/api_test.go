// +build unittest

package manager

import . "gopkg.in/check.v1"

type apiSuite struct {
}

var (
	_ = Suite(&apiSuite{})
)

// some POST handlers have static error checks, this test validates those
func (s *apiSuite) TestPostHandlerErrorCase(c *C) {
	m := Manager{}
	tests := map[string]struct {
		arg      *APIRequest
		cb       postCallback
		exptdErr error
	}{
		"monitor-event-invalid-name": {
			cb: m.monitorEvent,
			arg: &APIRequest{
				Event: MonitorEvent{
					Name: "foo",
				},
			},
			exptdErr: errInvalidEventName("foo"),
		},
		"monitor-event-empty-name": {
			cb: m.monitorEvent,
			arg: &APIRequest{
				Event: MonitorEvent{},
			},
			exptdErr: errInvalidEventName(""),
		},
	}

	for key, test := range tests {
		err := test.cb(test.arg)
		c.Assert(err, NotNil)
		c.Assert(err.Error(), Equals, test.exptdErr.Error(), Commentf("key: %s", key))
	}
}

// some Get handlers have static error checks, this test validates those
func (s *apiSuite) TestGetHandlerErrorCase(c *C) {
	m := Manager{}
	tests := map[string]struct {
		arg      *APIRequest
		cb       getCallback
		exptdErr error
	}{
		"job-invalid-label": {
			cb: m.jobGet,
			arg: &APIRequest{
				Job: "foo",
			},
			exptdErr: errInvalidJobLabel("foo"),
		},
		"job-empty-label": {
			cb:       m.jobGet,
			arg:      &APIRequest{},
			exptdErr: errInvalidJobLabel(""),
		},
		"job-non-existent": {
			cb: m.jobGet,
			arg: &APIRequest{
				Job: "active",
			},
			exptdErr: errJobNotExist("active"),
		},
	}

	for key, test := range tests {
		_, err := test.cb(test.arg)
		c.Assert(err, NotNil)
		c.Assert(err.Error(), Equals, test.exptdErr.Error(), Commentf("key: %s", key))
	}
}
