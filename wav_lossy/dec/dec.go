package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
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
	floatChunkSize := float64(chunkSize)
	var samplesPerChunk uint16
	if err := binary.Read(bufReader, binary.LittleEndian, &samplesPerChunk); err != nil {
		panic(fmt.Sprintf("error while reading samples per chunk: %v", err))
	}

	// numberOfCompleteChunks := numFrames / uint64(samplesPerChunk)
	// extraChunk := numFrames%uint64(samplesPerChunk) != 0
	channelCoefficients := make([][]float64, numChannels)
	for c := range channelCoefficients {
		channelCoefficients[c] = make([]float64, chunkSize)
	}
	channelValues := make([][]float64, numChannels)
	for c := range channelValues {
		channelValues[c] = make([]float64, samplesPerChunk)
	}
	outputBuffer := audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: int(numChannels),
			SampleRate:  int(samplingRate),
		},
		Data:           make([]int, samplesPerChunk*numChannels),
		SourceBitDepth: int(bitDepth),
	}
	// if numberOfChunks >= 1 {
	framesRemaining := numFrames
	for ; framesRemaining >= uint64(samplesPerChunk); framesRemaining -= uint64(samplesPerChunk) {
		for c, coefficients := range channelCoefficients {
			if err := binary.Read(bufReader, binary.LittleEndian, coefficients); err != nil {
				panic(fmt.Sprintf("error while reading channel %v coefficients: %v", c, err))
			}
			// fmt.Println(coefficients)
			for i := range channelValues[c] {
				channelValues[c][i] = coefficients[0] / floatChunkSize
			}
			for f, coefficient := range coefficients[1:] {
				for t := range samplesPerChunk {
					channelValues[c][t] += 2 * coefficient * math.Cos(float64(f+1)*(float64(t)+0.5)*math.Pi/floatChunkSize) / floatChunkSize
				}
			}
		}
		for c, values := range channelValues {
			for t, value := range values {
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
		for c, coefficients := range channelCoefficients {
			if err := binary.Read(bufReader, binary.LittleEndian, coefficients); err != nil {
				panic(fmt.Sprintf("error while reading channel %v coefficients: %v", c, err))
			}
			// fmt.Println(coefficients)
			for i := range channelValues[c] {
				channelValues[c][i] = coefficients[0] / floatChunkSize
			}
			for f, coefficient := range coefficients[1:] {
				for t := range samplesPerChunk {
					channelValues[c][t] += 2 * coefficient * math.Cos(float64(f+1)*(float64(t)+0.5)*math.Pi/floatChunkSize) / floatChunkSize
				}
			}
		}
		fmt.Println(channelValues)
		for c, values := range channelValues {
			for t := range framesRemaining {
				outputBuffer.Data[uint64(c)+uint64(t)*uint64(numChannels)] = int(math.Round(values[t]))
			}
		}
		fmt.Println(outputBuffer)
		if err := encoder.Write(&outputBuffer); err != nil {
			panic(fmt.Sprintf("error while writing to output file: %v", err))
		}
	}
}
