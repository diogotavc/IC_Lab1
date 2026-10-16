package main

import (
	"fmt"
	"os"

	"github.com/go-audio/wav"
)

func main() {
	audio1, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	defer audio1.Close()
	audio2, err := os.Open(os.Args[2])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	defer audio2.Close()

	decoder1 := wav.NewDecoder(audio1)
	decoder2 := wav.NewDecoder(audio2)
	buffer1, err := decoder1.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	buffer2, err := decoder2.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	numChannels := buffer1.Format.NumChannels
	if numChannels != buffer2.Format.NumChannels {
		panic("error: both files should have the same number of channels")
	}

	if len(buffer1.Data) != len(buffer2.Data) {
		panic("error: both files should have the same number of samples")
	}

	numFrames := len(buffer1.Data) / numChannels
	squaredError := make([]float64, numChannels)

	for f := range numFrames {
		for c := range numChannels {
			err := float64(buffer1.Data[f*numChannels+c]) - float64(buffer2.Data[f*numChannels+c])
			squaredError[c] += err * err
		}
	}

	meanSE := make([]float64, numChannels)
	for c := range numChannels {
		meanSE[c] = squaredError[c] / float64(numFrames)
	}

	var averageMSE float64
	for c := range numChannels {
		averageMSE += meanSE[c]
	}
	averageMSE = averageMSE / float64(numChannels)
}
