package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	flagDifferential uint8 = 1 << 0
	flagSlicing      uint8 = 1 << 1
)

func main() {
	if len(os.Args) < 2 {
		panic("input file not provided")
	}
	if len(os.Args) < 3 {
		panic("output file not provided")
	}

	inputFile, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0)
	if err != nil {
		panic(fmt.Sprintf("error while opening input file: %v", err))
	}
	defer inputFile.Close()
	bufReader := bufio.NewReader(inputFile)

	outputFile, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		panic(fmt.Sprintf("error while opening output file: %v", err))
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			panic(fmt.Sprintf("error while closing output file: %v", err))
		}
	}()

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
		panic(fmt.Sprintf("error while reading number of frames: %v", err))
	}
	var flags uint8
	if err := binary.Read(bufReader, binary.LittleEndian, &flags); err != nil {
		panic(fmt.Sprintf("error while reading flags: %v", err))
	}
	if flags&^(flagDifferential|flagSlicing) != 0 {
		panic(fmt.Sprintf("unknown flags: %08b", flags))
	}
	differential := flags&flagDifferential != 0
	slicing := flags&flagSlicing != 0
	fmt.Printf("Channels: %v\nBit depth: %v\nSampling rate: %v\nNumber of frames: %v\nDifferential: %v\nSlicing: %v\n",
		numChannels, bitDepth, samplingRate, numFrames, differential, slicing)
}
