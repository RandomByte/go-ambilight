#!/bin/bash
cd $GOPATH/src/github.com/RandomByte/go-ambilight
GOOS=linux GOARCH=arm GOARM=7 go build -o _test/goambi_raspi_build goambi.go && scp _test/goambi_raspi_build <pi hostname>:~/goambi/goambi
