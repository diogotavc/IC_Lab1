package main

import (
	"fmt"
	"os"
	"github.com/go-audio/wav"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	decoder := wav.NewDecoder(file)

	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	histogram := make(map[int]uint, 0)
	for _, value := range buffer.Data {
		histogram[value]++
	}
	fmt.Println(histogram)
}
