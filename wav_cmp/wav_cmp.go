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
	var squaredErrorAverage float64

	for f := range numFrames {
		var mix1, mix2 float64
		for c := range numChannels {
			a := float64(buffer1.Data[f*numChannels+c])
			b := float64(buffer2.Data[f*numChannels+c])

			err := a - b
			squaredError[c] += err * err

			mix1 += a
			mix2 += b
		}
		avgError := (mix1 - mix2) / float64(numChannels)
		squaredErrorAverage += avgError * avgError
	}

	meanSE := make([]float64, numChannels)
	for c := range numChannels {
		meanSE[c] = squaredError[c] / float64(numFrames)
	}
	avgMSE := squaredErrorAverage / float64(numFrames)

	for c := range numChannels {
		fmt.Printf("MSE (Channel %d): %f\n", c, meanSE[c])
	}
	fmt.Printf("MSE (Average): %f\n", avgMSE)
}
