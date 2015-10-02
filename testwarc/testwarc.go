package main

import (
        "warc"
        "os"
        "fmt"
)

func main() {
        infilename := os.Args[1]
        f, _ := os.Open(infilename)
        defer f.Close()
        wf, _ := warc.NewWARCFile(f)
        reader := wf.GetReader()
        count := 0
        reader.Iterate(func(wr *warc.WARCRecord, err error) {
                if err == nil {
                        count++
                }
        })
        fmt.Printf("Counted %v records", count)
}