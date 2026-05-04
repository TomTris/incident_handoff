package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const BufferSize = 500

func main() {
	buffer := make([]byte, BufferSize)
	f, err := os.Open("../README.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	totalBytes := 0
	content := ""

	for {
		readBytes, err := f.Read(buffer)

		// We must process these 2 lines first before checking EOF or err != nil
		// execute this command "go doc io.Reader" to know more about the correct behavior
		content += string(buffer[:readBytes])
		totalBytes += readBytes

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println(content)
	fmt.Print(totalBytes)
}
