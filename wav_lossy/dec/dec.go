package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/zikichombo/dsp/dct"
	// "github.com/zerfoo/float16"
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
	// bufWriter := bufio.NewWriter(outputFile)
	// defer func() {
	// 	if err := bufWriter.Flush(); err != nil {
	// 		panic(fmt.Sprintf("error while flushing to output file: %v", err))
	// 	}
	// }()

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
	fmt.Printf("Channels: %v\nBit depth: %v\nSampling rate: %v\nNumber of frames: %v\n", numChannels, bitDepth, samplingRate, numFrames)
	encoder := wav.NewEncoder(outputFile, int(samplingRate), int(bitDepth), int(numChannels), 1)
	defer func() {
		if err := encoder.Close(); err != nil {
			panic(fmt.Sprintf("error while closing encoder: %v", err))
		}
	}()

	var chunkSize uint16
	if err := binary.Read(bufReader, binary.LittleEndian, &chunkSize); err != nil {
		panic(fmt.Sprintf("error while reading chunk size: %v", err))
	}
	var samplesPerChunk uint16
	if err := binary.Read(bufReader, binary.LittleEndian, &samplesPerChunk); err != nil {
		panic(fmt.Sprintf("error while reading samples per chunk: %v", err))
	}
	if chunkSize > samplesPerChunk {
		panic(fmt.Sprintf("a chunk should generate at least %v samples, file has %v samples per chunk", chunkSize, samplesPerChunk))
	}

	channelBuffer := make([][]float64, numChannels)
	for c := range channelBuffer {
		channelBuffer[c] = make([]float64, samplesPerChunk)
	}
	inputBuffer := make([]float32, chunkSize)
	remainingZeros := make([]float64, samplesPerChunk - chunkSize)
	outputBuffer := audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: int(numChannels),
			SampleRate:  int(samplingRate),
		},
		Data:           make([]int, uint(samplesPerChunk)*uint(numChannels)),
		SourceBitDepth: int(bitDepth),
	}
	framesRemaining := numFrames
	transform := dct.New(int(samplesPerChunk))
	for ; framesRemaining >= uint64(samplesPerChunk); framesRemaining -= uint64(samplesPerChunk) {
		for c, buffer := range channelBuffer {
			if err := binary.Read(bufReader, binary.LittleEndian, inputBuffer); err != nil {
				panic(fmt.Sprintf("error while reading channel %v coefficients: %v", c, err))
			}
			for i, value := range inputBuffer {
				buffer[i] = float64(value)
			}
			copy(buffer[chunkSize:], remainingZeros)
			transform.Inv(buffer)
		}
		for c, buffer := range channelBuffer {
			for t, value := range buffer {
				outputBuffer.Data[uint64(c)+uint64(t)*uint64(numChannels)] = int(math.Round(value))
			}
		}
		if err := encoder.Write(&outputBuffer); err != nil {
			panic(fmt.Sprintf("error while writing to output file: %v", err))
		}
	}
	if framesRemaining > 0 {
		fmt.Println("frames remaining", framesRemaining)
		outputBuffer.Data = outputBuffer.Data[:framesRemaining*uint64(numChannels)]
		for c, buffer := range channelBuffer {
			if err := binary.Read(bufReader, binary.LittleEndian, inputBuffer); err != nil {
				panic(fmt.Sprintf("error while reading channel %v coefficients: %v", c, err))
			}
			for i, value := range inputBuffer {
				buffer[i] = float64(value)
			}
			copy(buffer[chunkSize:], remainingZeros)
			transform.Inv(buffer)
		}
		for c, buffer := range channelBuffer {
			for t, value := range buffer[:framesRemaining] {
				outputBuffer.Data[uint64(c)+uint64(t)*uint64(numChannels)] = int(math.Round(value))
			}
		}
		if err := encoder.Write(&outputBuffer); err != nil {
			panic(fmt.Sprintf("error while writing to output file: %v", err))
		}
	}
}
