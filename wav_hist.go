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

	if buffer.Format.NumChannels != 2 {
		panic("error: only stereo files supported")
	}

	midHistogram := make(map[int]uint, 0)
	sideHistogram := make(map[int]uint, 0)
	for i := 0; i < len(buffer.Data); i += 2 {
		left := buffer.Data[i]
		right := buffer.Data[i + 1]
		midHistogram[(left + right) / 2] += 1
		sideHistogram[(left - right) / 2] += 1
	}

	fmt.Println(midHistogram)
	fmt.Println(sideHistogram)
}
