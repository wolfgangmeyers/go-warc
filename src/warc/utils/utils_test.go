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
	c.Assert(string(part.Read(-1)), Equals, "")
	
	part = NewFilePart(strings.NewReader(s.text), 5)
	c.Assert(string(part.Read(-1)), Equals, "aaaa\n")
	
	part = NewFilePart(strings.NewReader(s.text), 10)
	c.Assert(string(part.Read(-1)), Equals, "aaaa\nbbbb\n")
	
	// try with large data
	part = NewFilePart(strings.NewReader(strings.Repeat("a", 10000)), 10)
	c.Assert(len(part.Read(-1)), Equals, 10)
}

func (s *FilePartSuite) TestReadWithSize(c *C) {
	part := NewFilePart(strings.NewReader(s.text), 10)
	c.Assert(string(part.Read(3)), Equals, "aaa")
	c.Assert(string(part.Read(3)), Equals, "a\nb")
	c.Assert(string(part.Read(3)), Equals, "bbb")
	c.Assert(string(part.Read(3)), Equals, "\n")
	c.Assert(string(part.Read(3)), Equals, "")
}

func (s *FilePartSuite) TestReadline(c *C) {
	part := NewFilePart(strings.NewReader(s.text), 11)
	c.Assert(string(part.ReadLine()), Equals, "aaaa\n")
	c.Assert(string(part.ReadLine()), Equals, "bbbb\n")
	c.Assert(string(part.ReadLine()), Equals, "c")
	c.Assert(string(part.ReadLine()), Equals, "")
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
