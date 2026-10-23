package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"

	// "github.com/zerfoo/float16"
)

func main() {
	if len(os.Args) < 2 {
		panic("no input file path given")
	}
	file, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0)
	if err != nil {
		panic(fmt.Sprintf("error while opening file: %v", err))
	}
	defer file.Close()
	bufReader := bufio.NewReader(file)

	var encodedNumChannels uint8
	if err := binary.Read(bufReader, binary.LittleEndian, &encodedNumChannels); err != nil {
		panic(fmt.Sprintf("error while reading number of channels: %v", err))
	}
	numChannels := uint16(encodedNumChannels) + 1
	var encodedBitDepth uint8
	if err := binary.Read(bufReader, binary.LittleEndian, &encodedBitDepth); err != nil {
		panic(fmt.Sprintf("error while reading bit depth: %v", err))
	}
	bitDepth := uint16(encodedBitDepth) + 1
	var samplingRate uint32
	if err := binary.Read(bufReader, binary.LittleEndian, &samplingRate); err != nil {
		panic(fmt.Sprintf("error while reading sampling rate: %v", err))
	}
	var numFrames uint64
	if err := binary.Read(bufReader, binary.LittleEndian, &numFrames); err != nil {
		panic(fmt.Sprintf("error while reading number of frame: %v", err))
	}
	var chunkSize uint16
	if err := binary.Read(bufReader, binary.LittleEndian, &chunkSize); err != nil {
		panic(fmt.Sprintf("error while reading chunk size: %v", err))
	}
	var samplesPerChunk uint16
	if err := binary.Read(bufReader, binary.LittleEndian, &samplesPerChunk); err != nil {
		panic(fmt.Sprintf("error while reading samples per chunk: %v", err))
	}
	fmt.Printf("Channels: %v\nBit depth: %v\nSampling rate: %v\nNumber of frames: %v\nNumber of coefficients in chunk: %v\nNumber of samples per chunks: %v\n", numChannels, bitDepth, samplingRate, numFrames, chunkSize, samplesPerChunk)

	numberOfChunks := numFrames / uint64(samplesPerChunk)
	if numFrames%uint64(samplesPerChunk) != 0 {
		numberOfChunks += 1
	}

	channelCoefficients := make([][]float32, numChannels)
	for c := range channelCoefficients {
		channelCoefficients[c] = make([]float32, chunkSize)
	}
	for range numberOfChunks {
		for c := range channelCoefficients {
			if err := binary.Read(bufReader, binary.LittleEndian, channelCoefficients[c]); err != nil {
				panic(fmt.Sprintf("error while reading channel %v coefficients: %v", c, err))
			}
		}
		fmt.Println(channelCoefficients)
	}
}
