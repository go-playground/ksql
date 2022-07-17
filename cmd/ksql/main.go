package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-playground/ksql"
	"github.com/go-playground/pkg/v5/bytes"
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
		w := bufio.NewWriter(os.Stdout)
		enc := json.NewEncoder(w)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 0, 200*bytesext.KiB), 5*bytesext.MiB)
		for scanner.Scan() {
			result, err := ex.Calculate(scanner.Bytes())
			if err != nil {
				fmt.Fprintln(os.Stderr, "reading standard input:", err)
				return
			}
			if err := enc.Encode(result); err != nil {
				fmt.Fprintln(os.Stderr, "encoding result to standard output:", err)
				return
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
		if err = w.Flush(); err != nil {
			fmt.Fprintln(os.Stderr, "writing standard output:", err)
		}
	} else {
		input = []byte(args[1])
		result, err := ex.Calculate(input)
		if err != nil {
			usage()
			return
		}
		enc := json.NewEncoder(os.Stderr)
		if err := enc.Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, "encoding result to standard output:", err)
			return
		}
	}
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
