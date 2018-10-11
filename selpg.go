package main

import (
	"bufio"
	"bytes"
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"os"
	"os/exec"
)

/* special type determined for saving arguments */
type SelpgArgs struct {
	startPage int
	endPage int
	pageLength int
	formFeed bool
	destination string
	inputFile string
}

/* check arguments */
func processArgs(psa *SelpgArgs , nonFlags []string, allArgs []string) {

	findL := func() bool {
		for i := 0; i < len(allArgs); i++ {
			if allArgs[i] == "--l" {
				return true
			}
		}
		return false
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	// check option arguments: --s--e --l --f
	switch {
	case psa.startPage < 1 :
		panic("selpg: Valid start page number does not exist.")
	case psa.endPage < 1 :
		panic("selpg: Valid end page number does not exist.")
	case psa.endPage < psa.startPage :
		panic("selpg: The end page number should be not smaller than the start page number.")
	case psa.pageLength < 1 :
		panic("selpg: Invalid page length.")
	case psa.formFeed && (psa.pageLength != 72 || findL()):
		panic("selpg: Only one way of delimiting pages can be chosen.")
	}

	// check non-option arguments: inputFile
	switch l := len(nonFlags); {
	case l > 1 :
		panic("selpg: There should be at most one input file.")
	case l == 1 :
		psa.inputFile = nonFlags[0]
	}
}

/* make output, error output and printing */
func processInput(psa *SelpgArgs){
	processError := func(err error){
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var reader *bufio.Reader
	var writer bytes.Buffer

	// judge which types of input to read, standard input or file
	if psa.inputFile == "" {
		reader = bufio.NewReader(os.Stdin)
	} else {
		file, err := os.Open(psa.inputFile)
		defer file.Close()
		reader = bufio.NewReader(file)
		if err != nil {
			processError(err)
		}
	}

	pageCounter := 1
	if psa.formFeed {	// delimit pages by form feeds
		for {
			page, err := reader.ReadString('\f')
			if err != nil && err != io.EOF{
				processError(err)
			}
			if psa.startPage <= pageCounter && pageCounter < psa.endPage{
				writer.WriteString(page)
			}
			pageCounter++
			if pageCounter == psa.endPage || err == io.EOF {
				break
			}
		}
	} else {			// delimit pages by a number of lines
		lineCounter := 1
		for {
			line, err := reader.ReadString('\n')
			if err != nil && err != io.EOF{
				processError(err)
			}
			if psa.startPage <= pageCounter && pageCounter < psa.endPage{
				writer.WriteString(line)
			}
			lineCounter++
			if lineCounter > psa.pageLength {
				pageCounter++
				lineCounter = 1
			}
			if pageCounter == psa.endPage || err == io.EOF {
				break
			}
		}
	}

	// if standard input is reading, does not stop reading until it ends
	if psa.inputFile == "" {
		for {
			_, err := reader.ReadByte()
			if err == io.EOF {
				break
			} else if err != nil {
				processError(err)
			}
		}
	}

	// output the result to standard output
	fmt.Fprint(os.Stdout, writer.String())

	// if the destination is specified, send the result to a printer
	if psa.destination != "" {
		cmd := exec.Command("lp", "-d ", psa.destination)
		in, err := cmd.StdinPipe()
		if err == nil {
			processError(err)
		}
		fmt.Fprint(in, writer.String())
		defer in.Close()
		out, err := cmd.CombinedOutput()
		if err == nil {
			processError(err)
		}
		fmt.Println(out)
	}
}

func Usage(){
	fmt.Println("Usage of selpg: selpg --s startPage --e endPage [ --f | --l pageLength] [--d destination] [inputFile]\n",
		"\t--s int\t\tfirst page to be printed\n",
		"\t--e int\t\tfirst page not printed after several continuous printed pages\n",
		"\t--f\t\tdelimit pages by form feeds\n",
		"\t--l int\t\tnumber of lines per page (default 72)\n",
		"\t--d string\tdestination printer")
}

func main() {
	var sa SelpgArgs

	flag.Usage = Usage

	flag.IntVar(&sa.startPage, "s", -1,"first page to be printed")
	flag.IntVar(&sa.endPage, "e",-1,"first page not printed")
	flag.IntVar(&sa.pageLength, "l",72,"number of lines per page")
	flag.BoolVar(&sa.formFeed, "f",false,"delimit pages by form feeds")
	flag.StringVar(&sa.destination, "d", "", "destination printer")

	flag.Parse()

	processArgs(&sa, flag.Args(), os.Args[1:])
	processInput(&sa)
}