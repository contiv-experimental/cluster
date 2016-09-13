// +build unittest

package manager

import (
	"bytes"

	. "gopkg.in/check.v1"
)

type MultiWriterSuite struct {
}

var _ = Suite(&MultiWriterSuite{})

func (s *MultiWriterSuite) TestMultiWrite(c *C) {
	mw := &MultiWriter{}

	var buf1, buf2 bytes.Buffer
	mw.Add(&buf1)
	mw.Add(&buf2)
	c.Assert(len(mw.writers), Equals, 2)

	testStr := "foo bar"
	_, _ = mw.Write([]byte(testStr))
	c.Assert(buf1.String(), Equals, testStr)
	c.Assert(buf1.Bytes(), DeepEquals, buf2.Bytes())
}
