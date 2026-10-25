package main

import (
	"fmt"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const blockSize = 8192

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
	if err := decoder1.FwdToPCM(); err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	if err := decoder2.FwdToPCM(); err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	if decoder1.BitDepth != decoder2.BitDepth || decoder1.SampleRate != decoder2.SampleRate {
		panic("error: files must have the same bit depth and sample rate")
	}
	format := decoder1.Format()
	numChannels := format.NumChannels
	if numChannels != decoder2.Format().NumChannels {
		panic("error: both files should have the same number of channels")
	}
	if decoder1.PCMLen() != decoder2.PCMLen() {
		panic("error: both files should have the same number of samples")
	}

	bufferLen := blockSize * numChannels
	buf1 := &audio.IntBuffer{
		Format:         format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: int(decoder1.BitDepth),
	}
	buf2 := &audio.IntBuffer{
		Format:         format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: int(decoder2.BitDepth),
	}

	squaredError := make([]float64, numChannels)
	maxError := make([]float64, numChannels)
	energy := make([]float64, numChannels)
	var squaredErrorMono, maxErrorMono, energyMono float64
	var numFrames int64

	invCh := 1 / float64(numChannels)

	for {
		var n2 int
		var e2 error
		done := make(chan struct{})
		go func() {
			n2, e2 = decoder2.PCMBuffer(buf2)
			close(done)
		}()
		n1, e1 := decoder1.PCMBuffer(buf1)
		<-done
		if n1 != n2 {
			panic("error: both files should have the same number of samples")
		}
		if n1 == 0 {
			break
		}

		frames := n1 / numChannels
		a, b := buf1.Data[:frames*numChannels], buf2.Data[:frames*numChannels]

		for i := 0; i < len(a); i += numChannels {
			var monoError, monoOriginal float64
			fa, fb := a[i:i+numChannels], b[i:i+numChannels]
			for c := range fa {
				x := float64(fa[c])
				diff := x - float64(fb[c])
				monoError += diff
				monoOriginal += x
				squaredError[c] += diff * diff
				energy[c] += x * x
				if m := math.Abs(diff); m > maxError[c] {
					maxError[c] = m
				}
			}
			monoError *= invCh
			monoOriginal *= invCh
			energyMono += monoOriginal * monoOriginal
			squaredErrorMono += monoError * monoError
			if m := math.Abs(monoError); m > maxErrorMono {
				maxErrorMono = m
			}
		}
		numFrames += int64(frames)

		if n1 < bufferLen || e1 != nil || e2 != nil {
			break
		}
	}

	for c := 0; c < numChannels; c++ {
		fmt.Printf("\033[1mChannel %d\033[0m\n", c)
		fmt.Printf("MSE: %f\n", squaredError[c]/float64(numFrames))
		fmt.Printf("Max Error: %f\n", maxError[c])
		fmt.Printf("SNR: %f\n", 10*math.Log10(energy[c]/squaredError[c]))
	}
	fmt.Printf("\033[1mMono\033[0m\n")
	fmt.Printf("MSE: %f\n", squaredErrorMono/float64(numFrames))
	fmt.Printf("Max Error: %f\n", maxErrorMono)
	fmt.Printf("SNR: %f\n", 10*math.Log10(energyMono/squaredErrorMono))
}
