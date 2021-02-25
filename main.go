package main

import "io"

func main() {
	inReader, inWriter := io.Pipe()
	inReader.Close()
	inWriter.Close()
	inReader.Close()
}
