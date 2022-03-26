package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-playground/ksql"
)

func main() {
	args := os.Args[1:]
	isPipe := isInputFromPipe()
	if (len(args) < 2 && !isPipe) || (len(args) < 1 && isPipe) {
		usage()
		return
	}

	ex, err := ksql.Parse([]byte(args[0]))
	if err != nil {
		usage()
		return
	}

	var input []byte

	if isPipe {
		input, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			usage()
			return
		}
	} else {
		input = []byte(args[1])
	}

	result, err := ex.Calculate(input)
	if err != nil {
		usage()
		return
	}
	fmt.Println(result)
}

func usage() {
	fmt.Println("ksql <expression> <json>")
	fmt.Println("or")
	fmt.Println("echo '{{}}' | ksql <expression> -")
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
