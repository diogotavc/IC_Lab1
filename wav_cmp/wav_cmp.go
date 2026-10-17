package main

import (
	"fmt"
	"math"
	"os"

	"github.com/go-audio/wav"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s original.wav modified.wav\n", os.Args[0])
		os.Exit(1)
	}

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
	if decoder1.BitDepth != decoder2.BitDepth || decoder1.SampleRate != decoder2.SampleRate {
		panic("error: files must have the same bit depth and sample rate")
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
	var squaredErrorMono float64

	maxError := make([]float64, numChannels)
	var maxErrorMono float64

	energy := make([]float64, numChannels)
	var energyMono float64

	for f := range numFrames {
		var monoError, monoOriginal float64
		for c := range numChannels {
			a := float64(buffer1.Data[f*numChannels+c])
			b := float64(buffer2.Data[f*numChannels+c])

			diff := a - b
			monoError += diff
			monoOriginal += a
			squaredError[c] += diff * diff
			energy[c] += a * a

			if m := math.Abs(diff); m > maxError[c] {
				maxError[c] = m
			}
		}
		monoError /= float64(numChannels)
		monoOriginal /= float64(numChannels)
		energyMono += monoOriginal * monoOriginal
		squaredErrorMono += monoError * monoError
		if m := math.Abs(monoError); m > maxErrorMono {
			maxErrorMono = m
		}
	}

	meanSE := make([]float64, numChannels)
	snRatio := make([]float64, numChannels)
	for c := range numChannels {
		meanSE[c] = squaredError[c] / float64(numFrames)
		snRatio[c] = 10 * math.Log10(energy[c]/squaredError[c])
	}
	mseMono := squaredErrorMono / float64(numFrames)
	snrMono := 10 * math.Log10(energyMono/squaredErrorMono)

	for c := range numChannels {
		fmt.Printf("\033[1mChannel %d\033[0m\n", c)
		fmt.Printf("MSE: %f\n", meanSE[c])
		fmt.Printf("Max Error: %f\n", maxError[c])
		fmt.Printf("SNR: %f\n", snRatio[c])
	}

	fmt.Printf("\033[1mMono\033[0m\n")
	fmt.Printf("MSE: %f\n", mseMono)
	fmt.Printf("Max Error: %f\n", maxErrorMono)
	fmt.Printf("SNR: %f\n", snrMono)
}
