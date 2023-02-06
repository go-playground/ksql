package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/go-playground/ksql"
	"github.com/go-playground/pkg/v5/bytes"
)

func main() {

	var outputOriginal bool
	flag.BoolVar(&outputOriginal, "o", false, "Indicates if the original data will be output after applying the expression. The results of the expression MUST be a boolean otherwise the output will be ignored.")
	flag.Usage = usage
	flag.Parse()

	isPipe := isInputFromPipe()
	if (flag.NArg() < 2 && !isPipe) || (flag.NArg() < 1 && isPipe) {
		flag.Usage()
		return
	}

	ex, err := ksql.Parse([]byte(flag.Arg(0)))
	if err != nil {
		flag.Usage()
		return
	}

	var input []byte
	w := bufio.NewWriter(os.Stdout)

	if isPipe {
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Buffer(make([]byte, 0, 200*bytesext.KiB), 5*bytesext.MiB)

		if outputOriginal {
			for scanner.Scan() {
				input := scanner.Bytes()
				result, err := ex.Calculate(input)
				if err != nil {
					fmt.Fprintln(os.Stderr, "reading standard input:", err)
					return
				}
				if result, ok := result.(bool); ok && result {
					_, err := w.Write(input)
					if err != nil {
						fmt.Fprintln(os.Stderr, "writing standard output:", err)
					}
					err = w.WriteByte('\n')
					if err != nil {
						fmt.Fprintln(os.Stderr, "writing standard output:", err)
					}
				}
			}
		} else {
			enc := json.NewEncoder(w)
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
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
		if err = w.Flush(); err != nil {
			fmt.Fprintln(os.Stderr, "writing standard output:", err)
		}
	} else {
		input = []byte(flag.Arg(1))
		result, err := ex.Calculate(input)
		if err != nil {
			flag.Usage()
			return
		}
		if outputOriginal {
			if result, ok := result.(bool); ok && result {
				_, err := w.Write(input)
				if err != nil {
					fmt.Fprintln(os.Stderr, "writing standard output:", err)
				}
				err = w.WriteByte('\n')
				if err != nil {
					fmt.Fprintln(os.Stderr, "writing standard output:", err)
				}
			}
		} else {
			enc := json.NewEncoder(w)
			if err := enc.Encode(result); err != nil {
				fmt.Fprintln(os.Stderr, "encoding result to standard output:", err)
				return
			}
		}
		if err = w.Flush(); err != nil {
			fmt.Fprintln(os.Stderr, "writing standard output:", err)
		}
	}
}

func usage() {
	fmt.Println("ksql [OPTIONS] <EXPRESSION> [DATA]")
	flag.PrintDefaults()
}

func isInputFromPipe() bool {
	fileInfo, _ := os.Stdin.Stat()
	return fileInfo.Mode()&os.ModeCharDevice == 0
}
