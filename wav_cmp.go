package main

import (
	"fmt"
	"os"

	"github.com/go-audio/wav"
)

func main() {
	modifiedAudioFile, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	originalAudioFile, err := os.Open(os.Args[2])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	decoderModified := wav.NewDecoder(modifiedAudioFile)
	decoderOriginal := wav.NewDecoder(originalAudioFile)
	bufferModified, err := decoderModified.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	bufferOriginal, err := decoderOriginal.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	numChannels := bufferModified.Format.NumChannels
	if numChannels != bufferOriginal.Format.NumChannels {
		panic("error: both files should have the same number of channels")
	}

	runningTotal := 0.0
	for i := 0; i < len(bufferModified.Data); i += numChannels {
		totalModified := 0.0
		totalOriginal := 0.0
		for j := 0; j < numChannels; j += 1 {
			totalModified += float64(bufferModified.Data[i+j])
			totalOriginal += float64(bufferOriginal.Data[i+j])
		}
		averageModified := totalModified / float64(numChannels)
		averageOriginal := totalOriginal / float64(numChannels)
		diff := averageModified - averageOriginal
		runningTotal += diff * diff
	}

	fmt.Println("Mean squared error:", runningTotal*float64(numChannels)/float64(len(bufferModified.Data)))
}
