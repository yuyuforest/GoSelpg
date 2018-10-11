package main

import (
	"bytes"
	"io/ioutil"
	"strconv"
	"syscall"
)

/* generate files used to test selpg */
func main() {
	var writer1 bytes.Buffer
	var writer2 bytes.Buffer

	// make a file with 150 lines and no form feeds
	for i := 1; i < 151; i++ {
		writer1.WriteString(strconv.Itoa(i))
		writer1.WriteString("\n")
	}

	// make a file with 5 pages
	for i := 1; i < 6; i++ {
		writer2.WriteString(strconv.Itoa(i))
		writer2.WriteString("\f")
	}

	ioutil.WriteFile("test-pageLength", writer1.Bytes(), syscall.O_CREAT)
	ioutil.WriteFile("test-formFeed", writer2.Bytes(), syscall.O_CREAT)
}
