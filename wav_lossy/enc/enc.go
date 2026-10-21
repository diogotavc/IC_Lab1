package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/zikichombo/dsp/dct"
	// "github.com/zerfoo/float16"
)

const chunkSize = uint16(8192)
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

	bufferLen := chunkSize * uint16(numChannels)
	buffer := audio.IntBuffer{
		Format:         format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: int(decoder.BitDepth),
	}
	channelFloatBuffer := make([][]float64, numChannels)
	for c := range numChannels {
		channelFloatBuffer[c] = make([]float64, chunkSize)
	}
	transform := dct.New(int(chunkSize))
	for {
		samplesRead, err := decoder.PCMBuffer(&buffer)
		if err != nil {
			panic(fmt.Sprintf("error while reading PCM buffer: %v", err))
		}
		if samplesRead == 0 {
			break
		}
		for i := range uint(chunkSize) {
			for c := range uint(numChannels) {
				channelFloatBuffer[c][i] = float64(buffer.Data[i*uint(numChannels)+c])
			}
		}
		for _, buffer := range channelFloatBuffer {
			transform.Do(buffer)
			if err := binary.Write(writer, binary.LittleEndian, buffer); err != nil {
				panic(fmt.Sprintf("error writing to output: %v", err))
			}
		}
		if uint(samplesRead) < uint(bufferLen) {
			break
		}
	}
}
