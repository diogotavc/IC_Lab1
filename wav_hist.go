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
	bin_size := 1 << bin_exp
	bin_mask := bin_size - 1

	decoder := wav.NewDecoder(file)

	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	if decoder.BitDepth > 64 {
		panic("bit depths larger than 64 are not supported")
	}
	minValue := (-1 << (decoder.BitDepth - 1))
	maxValue := -(minValue + 1)

	outputFiles := make([]*os.File, buffer.Format.NumChannels)
	histograms := make([]map[int]uint, buffer.Format.NumChannels)
	for n := range buffer.Format.NumChannels {
		file, err := os.OpenFile(fmt.Sprint("channelHistogram", n, ".csv"), os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0o666)
		if err != nil {
			panic(fmt.Sprintf("error: %v", err))
		}
		outputFiles[n] = file
		histograms[n] = make(map[int]uint, 0)
	}
	defer func() {
		for _, file := range outputFiles {
			if err := file.Close(); err != nil {
				panic(fmt.Sprintf("error at closing: %v", err))
			}
		}
	}()

	if buffer.Format.NumChannels == 2 {
		midHistogramFile, err := os.OpenFile("midHistogram.csv", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0o666)
		if err != nil {
			panic(fmt.Sprint("error: could not open MID channel histogram file: ", err))
		}
		defer func() {
			if err := midHistogramFile.Close(); err != nil {
				panic(fmt.Sprint("error: could not close MID channel histogram file: ", err))
			}
		}()
		sideHistogramFile, err := os.OpenFile("sideHistogram.csv", os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0o666)
		if err != nil {
			panic(fmt.Sprint("error: could not open SIDE channel histogram file: ", err))
		}
		defer func() {
			if err := sideHistogramFile.Close(); err != nil {
				panic(fmt.Sprint("error: could not close SIDE channel histogram file: ", err))
			}
		}()

		midHistogram := make(map[int]uint, 0)
		sideHistogram := make(map[int]uint, 0)
		for i := 0; i < len(buffer.Data); i += 2 {
			left := buffer.Data[i]
			right := buffer.Data[i + 1]
			midHistogram[((left + right) / 2) &^ bin_mask] += 1
			sideHistogram[((left - right) / 2) &^ bin_mask] += 1
		}

		value := minValue
		for {
			if _, err := fmt.Fprintf(midHistogramFile, "%d,%d\n", value, midHistogram[value]); err != nil {
				panic(fmt.Sprint("error: could not write to MID channel histogram file: ", err))
			}

			if _, err := fmt.Fprintf(sideHistogramFile, "%d,%d\n", value, sideHistogram[value]); err != nil {
				panic(fmt.Sprint("error: could not write to SIDE channel histogram file: ", err))
			}

			if maxValue - value < bin_size {
				break
			}
			value += bin_size
		}
	}

	for n := 0; n < len(buffer.Data); n += buffer.Format.NumChannels {
		for c := range buffer.Format.NumChannels {
			histograms[c][buffer.Data[n + c] &^ bin_mask] += 1
		}
	}
	for c := range buffer.Format.NumChannels {
		file := outputFiles[c]
		value := minValue
		for {
			if _, err := fmt.Fprintf(file, "%d,%d\n", value, histograms[c][value]); err != nil {
				panic(fmt.Sprintf("error: could not write to channel %v histogram file: %v", c, err))
			}
			if maxValue - value < bin_size {
				break
			}
			value += bin_size
		}
	}
}
