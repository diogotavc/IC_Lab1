package main

import (
	"fmt"
	"os"

	"github.com/go-audio/wav"
)

func main() {
	fileAudio, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	fileOriginal, err := os.Open(os.Args[2])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	decoderAudio := wav.NewDecoder(fileAudio)
	decoderOriginal := wav.NewDecoder(fileOriginal)
	bufferAudio, err := decoderAudio.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	bufferOriginal, err := decoderOriginal.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	numChannelsAudio := bufferAudio.Format.NumChannels
	if numChannelsAudio != bufferOriginal.Format.NumChannels {
		panic("error: both files should have the same number of channels")
	}

	runningTotal := 0.0
	for i := 0; i < len(bufferAudio.Data); i += numChannelsAudio {
		averageAudio := 0.0
		averageOriginal := 0.0
		for j := 0; j < numChannelsAudio; j += 1 {
			averageAudio += float64(bufferAudio.Data[i+j])
			averageOriginal += float64(bufferOriginal.Data[i+j])
		}
		averageAudio /= float64(numChannelsAudio)
		averageOriginal /= float64(numChannelsAudio)
		diff := averageOriginal - averageAudio
		runningTotal += diff * diff
	}

	fmt.Println("Mean squared error:", runningTotal*float64(numChannelsAudio)/float64(len(bufferAudio.Data)))
}
