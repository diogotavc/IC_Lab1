package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-audio/wav"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	bin_exp := 0
	if len(os.Args) >= 3 {
		bin_exp, err = strconv.Atoi(os.Args[2])
		if err != nil {
			panic(fmt.Sprintf("error: %v", err))
		}
	}
	bin_mask := 1 << bin_exp - 1

	decoder := wav.NewDecoder(file)

	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	if buffer.Format.NumChannels == 2 {
		midHistogram := make(map[int]uint, 0)
		sideHistogram := make(map[int]uint, 0)
		for i := 0; i < len(buffer.Data); i += 2 {
			left := buffer.Data[i] &^ bin_mask
			right := buffer.Data[i + 1] &^ bin_mask
			midHistogram[(left + right) / 2] += 1
			sideHistogram[(left - right) / 2] += 1
		}
		fmt.Println(midHistogram)
		fmt.Println(sideHistogram)
	}
	histograms := make([]map[int]uint, buffer.Format.NumChannels)
	for n := range buffer.Format.NumChannels {
		histograms[n] = make(map[int]uint, 0)
	}
	for n := 0; n < len(buffer.Data); n += buffer.Format.NumChannels {
		for c := range buffer.Format.NumChannels {
			histograms[c][buffer.Data[n + c] &^ bin_mask] += 1
		}
	}
	for n := range buffer.Format.NumChannels {
		fmt.Println(histograms[n])
	}
}
