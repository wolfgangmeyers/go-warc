package warc

import (
	"bytes"
	"compress/gzip"
	. "gopkg.in/check.v1"
	"testing"
	//	"sort"
	//	"strings"
)

func Test(t *testing.T) {
	TestingT(t)
}

type WARCHeaderSuite struct{}

var warcHeaderSuite = Suite(&WARCHeaderSuite{})

func (s *WARCHeaderSuite) TestAttrs(c *C) {
	h := NewWARCHeader(map[string]string{
		"WARC-Type":      "response",
		"WARC-Record-ID": "<record-1>",
		"WARC-Date":      "2000-01-02T03:04:05Z",
		"Content-Length": "10",
	})
	c.Assert(h.GetType(), Equals, "response")
	c.Assert(h.GetRecordId(), Equals, "<record-1>")
	c.Assert(h.GetDate(), Equals, "2000-01-02T03:04:05Z")
	c.Assert(h.GetContentLength(), Equals, 10)
}

func (s *WARCHeaderSuite) TestItemAccess(c *C) {
	h := NewWARCHeader(map[string]string{
		"WARC-Type":    "response",
		"X-New-Header": "42",
	})
	v, _ := h.Get("WARC-Type")
	c.Assert(v, Equals, "response")
	v, _ = h.Get("WARC-TYPE")
	c.Assert(v, Equals, "response")
	v, _ = h.Get("warc-type")
	c.Assert(v, Equals, "response")

	v, _ = h.Get("X-New-Header")
	c.Assert(v, Equals, "42")
	v, _ = h.Get("x-new-header")
	c.Assert(v, Equals, "42")
}

// TODO: test String(), InitDefaults() and new content types once read/write is implemented

func getSampleWarcRecord(numRecords int) []byte {
	text := "WARC/1.0\r\n" +
		"Content-Length: 10\r\n" +
		"WARC-Date: 2012-02-10T16:15:52Z\r\n" +
		"Content-Type: application/http; msgtype=response\r\n" +
		"WARC-Type: response\r\n" +
		"WARC-Record-ID: <urn:uuid:80fb9262-5402-11e1-8206-545200690126>\r\n" +
		"WARC-Target-URI: http://example.com/\r\n" +
		"\r\n" +
		"Helloworld" +
		"\r\n\r\n"
	buf := bytes.Buffer{}
	gzout := gzip.NewWriter(&buf)
	for i := 0; i < numRecords; i++ {
		gzout.Write([]byte(text))
		gzout.Flush()
		gzout.Reset(&buf)
	}
	
	return buf.Bytes()
}

type WARCReaderSuite struct{}

var warcReaderSuite = Suite(&WARCReaderSuite{})

func (s *WARCReaderSuite) TestReadHeader1(c *C) {
	reader := bytes.NewReader(getSampleWarcRecord(1))
	gzreader, err := gzip.NewReader(reader)
	c.Assert(err, Equals, nil)
	warcReader := NewWARCReader(reader, gzreader)
	record, err := warcReader.ReadRecord()
	c.Assert(err, Equals, nil)
	header := record.GetHeader()
	c.Assert(header.GetDate(), Equals, "2012-02-10T16:15:52Z")
	c.Assert(header.GetRecordId(), Equals, "<urn:uuid:80fb9262-5402-11e1-8206-545200690126>")
	c.Assert(header.GetType(), Equals, "response")
	c.Assert(header.GetContentLength(), Equals, 10)
}

func (s *WARCReaderSuite) TestEOF(c *C) {
	reader := bytes.NewReader(getSampleWarcRecord(1))
	gzreader, err := gzip.NewReader(reader)
	c.Assert(err, Equals, nil)
	warcReader := NewWARCReader(reader, gzreader)
	record, err := warcReader.ReadRecord()
	// read again and hit EOF
	record, err = warcReader.ReadRecord()
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "EOF")
	c.Assert(record, Equals, (*WARCRecord)(nil))
}

func (s *WARCReaderSuite) TestReadRecord(c *C) {
	reader := bytes.NewReader(getSampleWarcRecord(1))
	gzreader, err := gzip.NewReader(reader)
	c.Assert(err, Equals, nil)
	warcReader := NewWARCReader(reader, gzreader)
	record, err := warcReader.ReadRecord()
	c.Assert(record, NotNil)
	payload := record.GetPayload()
	lines := []string{}
	payload.Iterate(func(line []byte) {
		lines = append(lines, string(line))
	})
	c.Assert(len(lines), Equals, 1)
	c.Assert(lines[0], Equals, "Helloworld")
}

func (s *WARCReaderSuite) TestReadMultipleRecords(c *C) {
	reader := bytes.NewReader(getSampleWarcRecord(5))
	gzreader, err := gzip.NewReader(reader)
	c.Assert(err, Equals, nil)
	warcReader := NewWARCReader(reader, gzreader)
	for i := 0; i < 5; i++ {
		record, err := warcReader.ReadRecord()
		c.Assert(err, IsNil)
		c.Assert(record, NotNil)
	}
}

type WARCFileSuite struct{}

var warcFileSuite = Suite(&WARCFileSuite{})

// just to give the WARC file something that can be closed
// borrowed from https://groups.google.com/forum/#!topic/golang-nuts/J-Y4LtdGNSw
type ClosingBuffer struct { 
    *bytes.Reader 
} 
func (cb *ClosingBuffer) Close() error { 
    // we don't actually have to do anything here,
    // since the buffer is just some data in memory 
    // and the error is initialized to no-error 
    return nil
} 

func (w *WARCFileSuite) TestRead(c *C) {
	reader := bytes.NewReader(getSampleWarcRecord(1))
	f, err := NewWARCFile(&ClosingBuffer{reader})
	c.Assert(err, IsNil)
	record, _ := f.ReadRecord()
	c.Assert(record, NotNil)
	record, _ = f.ReadRecord()
	c.Assert(record, IsNil)
}

// TODO: add tests for write gz and test long header when read/write is implemented

