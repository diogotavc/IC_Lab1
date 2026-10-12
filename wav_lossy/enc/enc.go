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

const chunkSize = uint16(128)
const samplesPerChunk = chunkSize
const floatChunkSize = float64(chunkSize)

func main() {
	if len(os.Args) < 2 {
		panic("no input file passed")
	} else if len(os.Args) < 3 {
		panic("no output file passed")
	}
	file, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0)
	if err != nil {
		panic(fmt.Sprintf("error while opening input file: %v", err))
	}
	defer func() {
		if err := file.Close(); err != nil {
			panic(fmt.Sprintf("error while closing input file: %v", err))
		}
	}()
	decoder := wav.NewDecoder(file)
	decoder.ReadInfo()
	decoder.FwdToPCM()
	numFrames := uint64(decoder.PCMLen()) / uint64(decoder.NumChans) * 8 / uint64(decoder.BitDepth)
	fmt.Println("number of frames", numFrames)
	format := decoder.Format()

	outputFile, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
	if err != nil {
		panic(fmt.Sprintf("error while opening output file: %v", err))
	}
	defer func() {
		if err := outputFile.Close(); err != nil {
			panic(fmt.Sprintf("error while closing output file: %v", err))
		}
	}()
	writer := bufio.NewWriter(outputFile)
	defer func() {
		if err := writer.Flush(); err != nil {
			panic(fmt.Sprintf("error while flushing output file: %v", err))
		}
	}()
	if format.NumChannels > 256 || format.NumChannels == 0 {
		panic(fmt.Sprintf("number of channels must be between [1; 256], but input file has %v channels", format.NumChannels))
	}
	numChannels := uint8(format.NumChannels)
	if err := binary.Write(writer, binary.LittleEndian, numChannels-1); err != nil {
		panic(fmt.Sprintf("error writing to output: %v", err))
	}
	if decoder.BitDepth > 256 || decoder.BitDepth == 0 {
		panic(fmt.Sprintf("number of channels must be between [1; 256], but input file has %v channels", numChannels))
	}
	bitDepth := uint8(decoder.BitDepth)
	if err := binary.Write(writer, binary.LittleEndian, bitDepth-1); err != nil {
		panic(fmt.Sprintf("error writing to output: %v", err))
	}
	if err := binary.Write(writer, binary.LittleEndian, decoder.SampleRate); err != nil {
		panic(fmt.Sprintf("error writing to output: %v", err))
	}
	if err := binary.Write(writer, binary.LittleEndian, numFrames); err != nil {
		panic(fmt.Sprintf("error writing to output: %v", err))
	}
	if err := binary.Write(writer, binary.LittleEndian, uint16(chunkSize)); err != nil {
		panic(fmt.Sprintf("error writing to output: %v", err))
	}
	if err := binary.Write(writer, binary.LittleEndian, samplesPerChunk); err != nil {
		panic(fmt.Sprintf("error while writing to output: %v", err))
	}

	cosineValuesLength := uint(chunkSize) * uint(chunkSize)
	cosineValues := make([]float64, cosineValuesLength)
	for f := range uint(chunkSize) {
		for t := range uint(chunkSize) {
			cosineValues[f*uint(chunkSize)+t] = math.Cos(float64(f) * (float64(t) + .5) * math.Pi / floatChunkSize)
		}
	}
	fmt.Println(cosineValues)

	bufferLen := chunkSize * uint16(numChannels)
	floatZeroes := make([]float64, bufferLen)
	buffer := audio.IntBuffer{
		Format:         format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: int(decoder.BitDepth),
	}
	channelCoefficients := make([][]float64, numChannels)
	for c := range numChannels {
		channelCoefficients[c] = make([]float64, chunkSize)
	}
	for {
		samplesRead, err := decoder.PCMBuffer(&buffer)
		// fmt.Println("read", samplesRead, "samples")
		if err != nil {
			panic(fmt.Sprintf("error while reading PCM buffer: %v", err))
		}
		if samplesRead == 0 {
			break
		}
		for f := range uint(chunkSize) {
			i := f * uint(chunkSize)
			for t := range uint(chunkSize) {
				// cosineValue := math.Cos(f * (float64(t) + 0.5) * math.Pi / floatChunkSize)
				offset := uint(t) * uint(numChannels)
				for c := range uint(numChannels) {
					channelCoefficients[c][f] += float64(buffer.Data[offset+c]) * cosineValues[i+t]
				}
			}
		}
		for _, coefficients := range channelCoefficients {
			// for _, coefficient := range coefficients {
			// 	if err := binary.Write(writer, binary.LittleEndian, float16.FromFloat64(coefficient)); err != nil {
			// 		panic(fmt.Sprintf("error writing to output: %v", err))
			// 	}
			// }
			if err := binary.Write(writer, binary.LittleEndian, coefficients); err != nil {
				panic(fmt.Sprintf("error writing to output: %v", err))
			}
		}
		// fmt.Println(channelCoefficients)

		// for f := range chunkSize {
		// 	for c := range numChannels {
		// 		for i := range chunkSize {
		// 			t := float64(f) * (float64(i) + 0.5) * math.Pi / float64(chunkSize)
		// 			coefficients[c][f] += float64(buffer.Data[i*numChannels+c]) * math.Cos(t)
		// 		}
		// 		coefficients[c][f] *= 2
		// 		coefficients[c][f] /= float64(chunkSize)
		// 	}
		// }
		// for c := range numChannels {
		// 	coefficients[c][0] /= 2
		// }
		// fmt.Println(coefficients)
		if uint(samplesRead) < uint(bufferLen) {
			break
		}
		for _, coefficients := range channelCoefficients {
			copy(coefficients, floatZeroes)
		}
	}
}
