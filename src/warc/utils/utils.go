package utils

import (
	"bytes"
	"io"
	"math"
	"strings"
	"errors"
//	"fmt"
)

// Provides map-like behavior with case-insensitive keys
type CIStringMap struct {
	m map[string]string
}

func NewCIStringMap() *CIStringMap {
	return &CIStringMap{m: map[string]string{}}
}

func (mm *CIStringMap) Get(key string) (string, bool) {
	result, exists := mm.m[strings.ToLower(key)]
	return result, exists
}

func (mm *CIStringMap) Set(key string, value string) {
	mm.m[strings.ToLower(key)] = value
}

func (mm *CIStringMap) Delete(key string) {
	delete(mm.m, strings.ToLower(key))
}

func (mm *CIStringMap) Update(m map[string]string) {
	for key, value := range m {
		mm.m[strings.ToLower(key)] = value
	}
}

func (mm *CIStringMap) Keys() []string {
	result := make([]string, len(mm.m))
	i := 0
	for key, _ := range mm.m {
		result[i] = key
		i++
	}
	return result
}

func (mm *CIStringMap) Items(callback func(string, string)) {
	for key, value := range mm.m {
		callback(key, value)
	}
}

// File interface over a part of a file
type FilePart struct {
	fileobj io.Reader
	length  int
	offset  int
	buf     []byte
}

// Creates a new FilePart object
func NewFilePart(fileobj io.Reader, length int) *FilePart {
	filePart := &FilePart{
		fileobj: fileobj,
		length:  length,
		offset:  0,
		buf:     []byte{},
	}
	return filePart
}

// reads up until the size specified
func (fp *FilePart) Read(size int) ([]byte, error) {
	if size == -1 {
		return fp.read(fp.length)
	} else {
		return fp.read(size)
	}
}

func (fp *FilePart) read(size int) ([]byte, error) {
	var content []byte
	if len(fp.buf) >= size {
		content = fp.buf[:size]
		fp.buf = fp.buf[size:]
	} else {
		size = int(math.Min(float64(size), float64(fp.length-fp.offset-len(fp.buf))))
		tmp := make([]byte, size)
		// if this read doesn't succeed, that's ok
		// because the buffer might still have content
		numRead, _ := fp.fileobj.Read(tmp)
//		if err != nil {
//			return nil, err
//		}
		tmp = tmp[:numRead]
		content = append(fp.buf, tmp...)
		fp.buf = []byte{}
	}
	fp.offset += len(content)
	if len(content) == 0 {
		return nil, errors.New("EOF")
	} else {
		return content, nil
	}
	
}

// backs up the reader to the beginning of the content
func (fp *FilePart) unread(content []byte) {
	fp.buf = append(content, fp.buf...)
	fp.offset -= len(content)
}

// Reads a single line of content
func (fp *FilePart) ReadLine() ([]byte, error) {
	result := []byte{}
	chunk, err := fp.read(1024)
	if err != nil {
		return nil, err
	}
	
	for findNewline(chunk) == -1 {
		result = append(result, chunk...)
		chunk, err = fp.read(1024)
		if err != nil && err.Error() == "EOF" {
			chunk = []byte{}
			break
		}
	}
	i := findNewline(chunk)
	if i != -1 {
		fp.unread(chunk[i+1:])
		chunk = chunk[:i+1]
	}
	result = append(result, chunk...)
	return result, nil
}

// Iterates and invokes the callback function for each line
func (fp *FilePart) Iterate(callback func([]byte)) {
	line, err := fp.ReadLine()
	if err != nil {
		return
	}
	for err == nil {
		callback(line)
		line, err = fp.ReadLine()
	}
}

func (fp *FilePart) GetReader() io.Reader {
	return fp.fileobj
}

func (fp *FilePart) GetLength() int {
	return fp.length
}

func findNewline(chunk []byte) int {
	return bytes.IndexByte(chunk, '\n')
}
