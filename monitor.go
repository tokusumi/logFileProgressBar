package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

const BUFSIZE = 4096

// Truncate a last line in fp. Reading file is operated.
func TruncateLastLine(fp *os.File) error {
	s, err := fp.Stat()
	if err != nil {
		return err
	}

	start := int(s.Size() - BUFSIZE)
	sep := []byte("\n")
	buf := make([]byte, BUFSIZE)
	lastSep := -1
	if start < 0 {
		start = 0
		buf = make([]byte, s.Size())
	}
loop:
	for {
		_, err := fp.ReadAt(buf, int64(start))
		if err != nil {
			return err
		}
		lastSep = bytes.LastIndex(buf, sep)
		if lastSep >= 0 {
			// last escape is found
			break loop
		} else {
			// not found.
			if start == 0 {
				// we reached at head
				break loop
			}

			// go to next loop
			start -= BUFSIZE
			buf = make([]byte, BUFSIZE)
			if start < 0 {
				start = 0
				buf = make([]byte, s.Size())
			}
		}
	}
	// truncate
	err = fp.Truncate(int64(start + lastSep + 1))
	if err != nil {
		return err
	}
	return nil
}

// trimCR drops a terminal \r and lines before it from the data.
func trimCR(data []byte) (bool, []byte) {
	length := len(data)
	if length <= 1 {
		return false, data
	}
	lastIdx := length - 1
	if data[length-2] == '\r' && data[length-1] == '\n' {
		if length == 2 {
			return false, data
		}
		// ignore CRLF
		lastIdx = length - 3
	}
	if i := bytes.LastIndexByte(data[:lastIdx], '\r'); i >= 0 {
		if length == 1 {
			return true, []byte{}
		}
		return true, data[i+1:]
	}
	return false, data
}

// ScanCRSeparatedLines is a split function for a Scanner that returns each line of
// text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is one optional carriage return followed
// by one mandatory newline. In regular expression notation, it is `\r?\n`.
// The last non-empty line of input will be returned even if it has no
// newline.
func ScanCRSeparatedLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	if i := bytes.IndexByte(data, '\r'); i >= 0 {
		// We doesn't have a newline-terminated line but have a line for animation
		return len(data), data, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Must set log file path")
		return
	}

	sc := bufio.NewScanner(os.Stdin)
	sc.Split(ScanCRSeparatedLines)

	dstFile, err := os.OpenFile(os.Args[1], os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		fmt.Println(err)
	}
	defer dstFile.Close()

	writer := bufio.NewWriter(dstFile)

	for sc.Scan() {
		v := sc.Bytes()

		includeCR, trimed_v := trimCR(v)
		if includeCR {
			// delete last line in file
			err := TruncateLastLine(dstFile)
			if err != nil {
				fmt.Println(err)
			}
		}
		// return newline to file
		_, err := writer.Write(trimed_v)
		if err != nil {
			fmt.Println(err)
		}

		// return to terminal
		// terminal can manage '\r'
		fmt.Print(string(v))

		writer.Flush()
	}
}
