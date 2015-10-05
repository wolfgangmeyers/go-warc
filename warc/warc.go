package warc

import (
	"bufio"
	"bytes"
	"compress/gzip"
//	"crypto/sha1"
//	"encoding/hex"
	"errors"
	"fmt"
//	"github.com/nu7hatch/gouuid" // needed for read/write
	"io"
	"regexp"
	"strconv"
	"strings"
//	"time" // needed for read/write
	"bitbucket.org/donutsinc/go-warc/warc/utils"
)

var CONTENT_TYPES map[string]string = map[string]string{
	"warcinfo": "application/warc-fields",
	"response": "application/http; msgtype=response",
	"request":  "application/http; msgtype=request",
	"metadata": "application/warc-fields",
}

var KNOWN_HEADERS map[string]string = map[string]string{
	"type":           "WARC-Type",
	"date":           "WARC-Date",
	"record_id":      "WARC-Record-ID",
	"ip_address":     "WARC-IP-Address",
	"target_uri":     "WARC-Target-URI",
	"warcinfo_id":    "WARC-Warcinfo-ID",
	"request_uri":    "WARC-Request-URI",
	"content_type":   "Content-Type",
	"content_length": "Content-Length",
}

var RE_VERSION *regexp.Regexp = regexp.MustCompile("WARC/(\\d+.\\d+)\r\n")
var RE_HEADER *regexp.Regexp = regexp.MustCompile("([a-zA-Z_\\-]+): *(.*)\r\n")
var SUPPORTED_VERSIONS map[string]bool = map[string]bool{"1.0": true}

//    The WARC Header object represents the headers of a WARC record.
//    It provides dictionary like interface for accessing the headers.
//
//    The following mandatory fields are accessible also as get/set methods.
//
//        * h.GetRecordId() == h.Get('WARC-Record-ID')
//        * h.GetContentLength() == h.Get("Content-Length") // converted to int
//        * h.GetDate() == h.Get("WARC-Date")
//        * h.GetType() == h.Get("WARC-Type")
//
//    :params headers: map[string]string of headers.
//    :params defaults: If true, important headers like WARC-Record-ID,
//                      WARC-Date, Content-Type and Content-Length are
//                      initialized to automatically if not already present.
//                      TODO: add this param back for read/write
type WARCHeader struct {
	version string
	*utils.CIStringMap
}

// TODO: restore 'defaults' arg for read/write
func NewWARCHeader(headers map[string]string/*, defaults bool*/) *WARCHeader {
	warcHeader := &WARCHeader{
		"WARC/1.0",
		utils.NewCIStringMap(),
	}
	warcHeader.Update(headers)
//	if defaults {
//		warcHeader.InitDefaults()
//	}
	return warcHeader
}

// TODO: restore this when the warc file becomes read-write
// Initializes important headers to default values, if not already specified.
//
// The WARC-Record-ID header is set to a newly generated UUID.
// The WARC-Date header is set to the current datetime.
// The Content-Type is set based on the WARC-Type header.
// The Content-Length is initialized to 0.
//func (wh *WARCHeader) InitDefaults() {
//	_, exists := wh.Get("WARC-Record-ID")
//	if !exists {
//		recordUUID, err := uuid.NewV4()
//		if err != nil {
//			panic(err)
//		}
//		wh.Set("WARC-Record-ID", fmt.Sprintf("<urn:uuid:%v>", recordUUID.String()))
//	}
//	_, exists = wh.Get("WARC-Date")
//	if !exists {
//		wh.Set("WARC-Date", time.Now().Format(time.RFC3339))
//	}
//	_, exists = wh.Get("Content-Type")
//	if !exists {
//		t := wh.GetType()
//		t, exists = CONTENT_TYPES[t]
//		if !exists {
//			t = "application/octet-stream"
//		}
//		wh.Set("Content-Type", t)
//	}
//}

// Writes this header to a file, in the format specified by WARC.
func (wh *WARCHeader) WriteTo(f io.Writer) {
	f.Write([]byte(wh.version + "\r\n"))
	wh.Items(func(name string, value string) {
		name = strings.Title(name)
		// Use standard forms for commonly used patterns
		name = strings.Replace(name, "Warc-", "WARC-", -1)
		name = strings.Replace(name, "-Ip-", "-IP-", -1)
		name = strings.Replace(name, "-Id", "-ID", -1)
		name = strings.Replace(name, "-Uri", "-URI", -1)
		f.Write([]byte(name + ": " + value + "\r\n"))
	})
	// Header ends with an extra CRLF
	f.Write([]byte("\r\n"))
}

// The Content-Length header as int.
func (wh *WARCHeader) GetContentLength() int {
	v, exists := wh.Get("Content-Length")
	if !exists {
		return 0
	}
	result, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		panic(err)
	}
	return int(result)
}

// The value of WARC-Record-ID header.
func (wh *WARCHeader) GetRecordId() string {
	v, _ := wh.Get("WARC-Record-ID")
	return v
}

// The value of WARC-Date header.
func (wh *WARCHeader) GetDate() string {
	v, _ := wh.Get("WARC-Date")
	return v
}

// The value of WARC-Type header.
func (wh *WARCHeader) GetType() string {
	t, _ := wh.Get("WARC-Type")
	return t
}

func (wh *WARCHeader) String() string {
	b := bytes.Buffer{}
	var f io.Writer = &b
	wh.WriteTo(f)
	return b.String()
}

// The WARCRecord object represents a WARC Record.
type WARCRecord struct {
	header  *WARCHeader
	payload *utils.FilePart
}

// Creates a new WARC record.
func NewWARCRecord(header *WARCHeader, payload *utils.FilePart, headers map[string]string) *WARCRecord {
	warcRecord := &WARCRecord{}
	if header == nil {
		// TODO: restore for read/write
		header = NewWARCHeader(headers/*, true*/)
	}
	warcRecord.header = header
	warcRecord.payload = payload
	return warcRecord
}

//func (wr *WARCRecord) computeDigest(payload []byte) string {
//	hash := sha1.Sum(payload)
//	return "sha1:" + hex.EncodeToString(hash[:20])
//}

// Record type
func (wr *WARCRecord) GetType() string {
	return wr.header.GetType()
}

// The value of the WARC-Target-URI header if the record is of type "response".
func (wr *WARCRecord) GetUrl() string {
	url, _ := wr.header.Get("WARC-Target-URI")
	return url
}

// The IP address of the host contacted to retrieve the content of this record.
// This value is available from the WARC-IP-Address header.
func (wr *WARCRecord) GetIpAddress() string {
	ipAddress, _ := wr.header.Get("WARC-IP-Address")
	return ipAddress
}

// UTC timestamp of the record.
func (wr *WARCRecord) GetDate() string {
	date, _ := wr.header.Get("WARC-Date")
	return date
}

func (wr *WARCRecord) GetChecksum() string {
	checksum, _ := wr.header.Get("WARC-Payload-Digest")
	return checksum
}

// Offset of this record in the warc file from which this record is read.
// TODO: not yet implemented. Currently hard-coded to -1
func (wr *WARCRecord) Offset() int {
	return -1
}

func (wr *WARCRecord) Get(name string) (string, bool) {
	v, exists := wr.header.Get(name)
	return v, exists
}

func (wr *WARCRecord) Set(name string, value string) {
	wr.header.Set(name, value)
}

func (wr *WARCRecord) GetHeader() *WARCHeader {
	return wr.header
}

func (wr *WARCRecord) GetPayload() *utils.FilePart {
	return wr.payload
}

//TODO: port the convenience method to create from http response.
// not sure yet how to port over the logic, or if it's

type WARCFile struct {
	filehandle io.ReadCloser
	filebuf *bufio.Reader
	gzipfile   *gzip.Reader
	reader     *WARCReader
}

// Creates a new WARCFile
// input should be a handle to a gzipped WARC file
func NewWARCFile(reader io.ReadCloser) (*WARCFile, error) {
	filebuf := bufio.NewReader(reader)
	gzipfile, err := gzip.NewReader(filebuf)
	if err != nil {
		return nil, err
	}
	// make sure to read each gzipped record separately
	gzipfile.Multistream(false)
	// keep a handle to underlying file so that it can be closed.
	wf := &WARCFile{
		filehandle: reader,
		filebuf: filebuf,
		gzipfile:   gzipfile,
		reader:     NewWARCReader(filebuf, gzipfile),
	}
	return wf, nil
}

func (wf *WARCFile) GetReader() *WARCReader {
	return wf.reader
}

func (wf *WARCFile) ReadRecord() (*WARCRecord, error) {
	return wf.reader.ReadRecord()
}

func (wf *WARCFile) Close() error {
	return wf.filehandle.Close()
}

type WARCReader struct {
	filehandle io.Reader
	gzipfile   *gzip.Reader
}

func NewWARCReader(filehandle io.Reader, gzipfile *gzip.Reader) *WARCReader {
	warcReader := &WARCReader{
		filehandle: filehandle,
		gzipfile:   gzipfile,
	}
	return warcReader
}

func (wr *WARCReader) ReadHeader(reader *bufio.Reader) (*WARCHeader, error) {
	versionLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	match := RE_VERSION.FindStringSubmatch(versionLine)
	if len(match) == 0 {
		return nil, errors.New(fmt.Sprintf("Bad version line: %v", versionLine))
	}
	version := match[1]
	supported := SUPPORTED_VERSIONS[version]
	if !supported {
		return nil, errors.New(fmt.Sprintf("Unsupported WARC version: %v", version))
	}
	headers := map[string]string{}
	for {
		line, err := reader.ReadString('\n')
//		fmt.Println("*** header line - " + line)
		if err != nil {
			return nil, err
		}
		if line == "\r\n" {
			break
		}
		match = RE_HEADER.FindStringSubmatch(line)
		if len(match) == 0 {
			return nil, errors.New(fmt.Sprintf("Bad header line: %v", line))
		}
		name, value := match[1], match[2]
		headers[name] = value
	}
	// TODO: restore for read/write
	return NewWARCHeader(headers/*, false*/), nil
}

func (wr *WARCReader) Expect(reader *bufio.Reader, expectedLine string, message string) error {
	line, err := reader.ReadString('\n')
	if err != nil {
//		fmt.Println(err)
		return err
	}
	if line != expectedLine {
		if message == "" {
			message = fmt.Sprintf("Expected %v, found %v", expectedLine, line)
		}
		return errors.New(message)
	}
	return nil
}

func (wr *WARCReader) ReadRecord() (*WARCRecord, error) {
	reader := bufio.NewReader(wr.gzipfile)
	header, err := wr.ReadHeader(reader)
	
	if err != nil && strings.Index(err.Error(), "EOF") > -1 {
		return nil, errors.New("EOF")
	}
	if err != nil {
		panic(err)
	}
	payload, err := utils.NewFilePart(reader, header.GetContentLength())
	if err != nil {
		panic(err)
	}
	// consume the footer from the previous record
	wr.Expect(reader, "\r\n", "")
	wr.Expect(reader, "\r\n", "")
	// the last call advances to the end of the gzip file
	wr.Expect(reader, "\r\n", "")
	// start reading the next record in the gzip file
	wr.gzipfile.Reset(wr.filehandle)
	wr.gzipfile.Multistream(false)
	
	if err != nil {
		return nil, err
	}
	record := NewWARCRecord(header, payload, map[string]string{})
	return record, nil
}

func (wr *WARCReader) Iterate(callback func(*WARCRecord, error)) {
	record, err := wr.ReadRecord()
	callback(record, err)
	for record != nil {
		record, err = wr.ReadRecord()
		callback(record, err)
	}
}
