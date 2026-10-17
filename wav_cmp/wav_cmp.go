package main

import (
	"fmt"
	"math"
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

	maxError := make([]float64, numChannels)
	var maxErrorAverage float64

	energy := make([]float64, numChannels)
	var energyAverage float64

	for f := range numFrames {
		var monoError, monoOriginal float64
		for c := range numChannels {
			a := float64(buffer1.Data[f*numChannels+c])
			b := float64(buffer2.Data[f*numChannels+c])

			err := a - b
			monoError += err
			monoOriginal += a
			squaredError[c] += err * err
			energy[c] += a * a

			if m := math.Abs(err); m > maxError[c] {
				maxError[c] = m
			}
		}
		monoError /= float64(numChannels)
		monoOriginal /= float64(numChannels)
		energyAverage += monoOriginal * monoOriginal
		squaredErrorAverage += monoError * monoError
		if m := math.Abs(monoError); m > maxErrorAverage {
			maxErrorAverage = m
		}
	}

	meanSE := make([]float64, numChannels)
	snRatio := make([]float64, numChannels)
	for c := range numChannels {
		meanSE[c] = squaredError[c] / float64(numFrames)
		snRatio[c] = 10 * math.Log10(energy[c]/squaredError[c])
	}
	avgMSE := squaredErrorAverage / float64(numFrames)
	avgSNR := 10 * math.Log10(energyAverage/squaredErrorAverage)

	for c := range numChannels {
		fmt.Printf("MSE (Channel %d): %f\n", c, meanSE[c])
		fmt.Printf("Max Error (Channel %d): %f\n", c, maxError[c])
		fmt.Printf("SNR (Channel %d): %f\n", c, snRatio[c])
	}
	fmt.Printf("MSE (Average): %f\n", avgMSE)
	fmt.Printf("Max Error (Average): %f\n", maxErrorAverage)
	fmt.Printf("SNR (Average): %f\n", avgSNR)
}
