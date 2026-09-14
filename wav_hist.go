package main

import (
	"fmt"
	"os"
	"flag"
	"github.com/go-audio/wav"
)

func main() {
	inputPath := flag.String("input path", "", "path to input file")
	// outputPath := flag.String("output path", nil, "path to output file")
	flag.Parse()

	if *inputPath == "" {
		panic(fmt.Sprintf("no input path given"))
	}

	file, err := os.Open(*inputPath)
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	decoder := wav.NewDecoder(file)

	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	histogram := make(map[int]int, 0)
	for _, value := range buffer.Data {
		histogram[value]++
	}
	fmt.Println(histogram)
}
