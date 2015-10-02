package main

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: gowarc <file>")
	}
	infilename := os.Args[1]
	infile, err := os.Open(infilename)
	if err != nil {
		panic(err)
	}
	defer infile.Close()
	reader, err := gzip.NewReader(infile)
	if err != nil {
		panic(err)
	}
	buffreader := bufio.NewReader(reader)
	for i := 0; i < 100; i++ {
		line, err := buffreader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		fmt.Println(strings.TrimSpace(line))
	}
}

