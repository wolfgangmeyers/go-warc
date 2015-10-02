go-warc: golang library to work with WARC files
============================================

go-warc is a golang port of the python warc library:

https://github.com/internetarchive/warc

Note that currently only reading of WARC files is supported. Writing
of WARC files may be implemented at some future point.

WARC (Web ARChive) is a file format for storing web crawls.

http://bibnum.bnf.fr/WARC/ 

This `warc` library makes it very easy to work with WARC files.::

    package main
    
    import (
        "fmt"
        "os"
        "bitbucket.org/donutsinc/go-warc/warc"
    )
    
    func main() {
        infilename := os.Args[1]
        f, err := os.Open(infilename)
        if err != nil {
            panic(err)
        }
        defer f.Close()
        wf, err := warc.NewWARCFile(f)
        if err != nil {
            panic(err)
        }
        reader := wf.GetReader()
        count := 0
        reader.Iterate(func(wr *warc.WARCRecord, err error) {
            if err == nil {
                count++
                fmt.Printf("Processed: %v - %v\n", wr.GetHeader().GetRecordId(), count)
                // you could do some other stuff with the record here
            }
        })
        fmt.Printf("Done!")
    }

Installing
--------
Make sure you have a working go environment. Instructions can be found here:

https://golang.org/doc/install

Currently go-warc builds against the standard go library with no external dependencies. To install the
go-warc library:

    go get bitbucket.org/donutsinc/go-warc/warc
    go get bitbucket.org/donutsinc/go-warc/warc/utils

Testing
-------
Navigate to the root of the project and run:

    $ go test warc warc/utils

Documentation
-------------

There isn't any yet. The original python
documentation of the warc library is available at http://warc.readthedocs.org/.
    
License
-------

This software is licensed under GPL v2. See [LICENSE](LICENSE.txt) file for details.
