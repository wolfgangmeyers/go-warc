package utils

import (
	. "gopkg.in/check.v1"
	"testing"
	"sort"
	"strings"
)

func Test(t *testing.T) {
	TestingT(t)
}

type CIStringMapSuite struct{}

var mSuite = Suite(&CIStringMapSuite{})

func (s *CIStringMapSuite) TestAll(c *C) {
	d := NewCIStringMap()
	d.Set("Foo", "1")
	result, exists := d.Get("foo")
	c.Assert(result, Equals, "1")
	c.Assert(exists, Equals, true)
	result, exists = d.Get("Foo")
	c.Assert(result, Equals, "1")
	c.Assert(exists, Equals, true)
	
	result, exists = d.Get("bar")
	c.Assert(result, Equals, "")
	c.Assert(exists, Equals, false)
	d.Set("BAR", "2")
	result, exists = d.Get("bar")
	c.Assert(result, Equals, "2")
	c.Assert(exists, Equals, true)
	
	keys := d.Keys()
	sort.Strings(keys)
	expectedKeys := []string{"bar", "foo"}
	c.Assert(len(keys), Equals, len(expectedKeys))
	
	for i := 0; i < 2; i++ {
		c.Assert(keys[i], Equals, expectedKeys[i])
	}
	
}

// update is new in the go version, and should be tested
func (s *CIStringMapSuite) TestUpdate(c *C) {
	d := NewCIStringMap()
	d.Update(map[string]string{"foo": "1", "BAR": "2"})
	keys := d.Keys()
	sort.Strings(keys)
	expectedKeys := []string{"bar", "foo"}
	c.Assert(len(keys), Equals, len(expectedKeys))
	for i := 0; i < len(keys); i++ {
		c.Assert(keys[i], Equals, expectedKeys[i])
	}
}

type FilePartSuite struct{
	text string
}

var fpSuite = Suite(&FilePartSuite{})

func (s *FilePartSuite) SetUpSuite(c *C) {
	s.text = strings.Join([]string{"aaaa", "bbbb", "cccc", "dddd", "eeee", "ffff"}, "\n")
}

func (s *FilePartSuite) TestRead(c *C) {
	part := NewFilePart(strings.NewReader(s.text), 0)
	data, _ := part.Read(-1)
	c.Assert(string(data), Equals, "")
	
	part = NewFilePart(strings.NewReader(s.text), 5)
	data, _ = part.Read(-1)
	c.Assert(string(data), Equals, "aaaa\n")
	
	part = NewFilePart(strings.NewReader(s.text), 10)
	data, _ = part.Read(-1)
	c.Assert(string(data), Equals, "aaaa\nbbbb\n")
	
	// try with large data
	part = NewFilePart(strings.NewReader(strings.Repeat("a", 10000)), 10)
	data, _ = part.Read(-1)
	c.Assert(len(data), Equals, 10)
}

func (s *FilePartSuite) TestReadWithSize(c *C) {
	part := NewFilePart(strings.NewReader(s.text), 10)
	data, _ := part.Read(3)
	c.Assert(string(data), Equals, "aaa")
	data, _ = part.Read(3)
	c.Assert(string(data), Equals, "a\nb")
	data, _ = part.Read(3)
	c.Assert(string(data), Equals, "bbb")
	data, _ = part.Read(3)
	c.Assert(string(data), Equals, "\n")
	data, _ = part.Read(3)
	c.Assert(string(data), Equals, "")
}

func (s *FilePartSuite) TestReadline(c *C) {
	part := NewFilePart(strings.NewReader(s.text), 11)
	data, _ := part.ReadLine()
	c.Assert(string(data), Equals, "aaaa\n")
	data, _ = part.ReadLine()
	c.Assert(string(data), Equals, "bbbb\n")
	data, _ = part.ReadLine()
	c.Assert(string(data), Equals, "c")
	data, _ = part.ReadLine()
	c.Assert(string(data), Equals, "")
}

func (s *FilePartSuite) TestIterate(c *C) {
	part := NewFilePart(strings.NewReader(s.text), 11)
	result := []string{}
	part.Iterate(func(line []byte) {
		result = append(result, string(line))
	})
	expected := []string{"aaaa\n", "bbbb\n", "c"}
	c.Assert(len(result), Equals, len(expected))
	for i := 0; i < len(result); i++ {
		c.Assert(result[i], Equals, expected[i])
	}
}
