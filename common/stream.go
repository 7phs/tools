package common

import (
	"bufio"
	"bytes"
	"io"
)

func MakeStream(reader io.Reader) chan string {
	if reader==nil {
		return nil
	}

	stream := make(chan string)
	bufReader := bufio.NewReader(reader)

	go func() {
		defer close(stream)

		for {
			if line, _, err := bufReader.ReadLine(); err!=nil {
				break
			} else {
				stream <- string(line) + "\n"
			}
		}
	}()

	return stream
}

func ReadAll(stream <-chan string) string {
	buffer := bytes.NewBufferString("")

	func() {
		writer := bufio.NewWriter(buffer)
		defer writer.Flush()

		for str := range stream {
			writer.WriteString(str)
		}
	}()

	return buffer.String()
}