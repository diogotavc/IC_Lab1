package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"slices"
	"flag"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/zikichombo/dsp/dct"
	// "github.com/zerfoo/float16"
)

const inlineZeros = 1

func main() {
	chunkSizeArg := flag.Int("block-size", 2048, "set encoder block size")
	flag.Parse()
	if *chunkSizeArg < 1 {
		panic("block size must be at least 1")
	} else if *chunkSizeArg > 32768 {
		panic("block size is at most 32768")
	}
	chunkSize := uint16(*chunkSizeArg)
	args := flag.Args()
	if len(args) < 1 {
		panic("no input file passed")
	} else if len(args) < 2 {
		panic("no output file passed")
	}
	file, err := os.OpenFile(args[0], os.O_RDONLY, 0)
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

	outputFile, err := os.OpenFile(args[1], os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
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
	if err := binary.Write(writer, binary.LittleEndian, chunkSize); err != nil {
		panic(fmt.Sprintf("error while writing to output: %v", err))
	}

	bufferLen := uint(chunkSize) * uint(numChannels)
	buffer := audio.IntBuffer{
		Format:         format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: int(decoder.BitDepth),
	}
	channelFloatBuffer := make([][]float64, numChannels)
	for c := range numChannels {
		channelFloatBuffer[c] = make([]float64, chunkSize)
	}
	outputBuffer := make([]int32, chunkSize)
	runLengthState := runLengthState{
		numberOfZeros: 0,
		nonZeros:      make([]int8, 0),
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
			for i, value := range buffer {
				outputBuffer[i] = int32(math.Round(value))
			}
			absMax := slices.Max(outputBuffer)
			if absMax < 0 {
				absMax = -absMax
			}
			absMin := slices.Min(outputBuffer)
			if absMin < 0 {
				absMin = -absMin
			}
			max := max(absMax, absMin)
			if err := binary.Write(writer, binary.LittleEndian, max); err != nil {
				panic(fmt.Sprintf("error while writing mutliplier value to output: %v\n", err))
			}
			floatMax := float64(max)
			for _, value := range outputBuffer {
				quantized := int8(math.Round(float64(value) * 127 / floatMax))
				// fmt.Println("real value:", buffer[i], "value:", value, "quantized:", quantized, "dequantized: ", float64(int32(quantized) * max)/127)
				runLengthState.receive(writer, quantized)
			}
			runLengthState.flush(writer)
		}
		if uint(samplesRead) < uint(bufferLen) {
			break
		}
	}
}

type runLengthState struct {
	numberOfZeros uint
	nonZeros      []int8
}

func (state *runLengthState) receive(writer *bufio.Writer, value int8) error {
	if value != 0 {
		if state.numberOfZeros > inlineZeros {
			if err := state.flush(writer); err != nil {
				return err
			}
		// } else if state.numberOfZeros == 1 {
		} else {
		 for range state.numberOfZeros {
				state.nonZeros = append(state.nonZeros, 0)
			}
			state.numberOfZeros = 0
		}
		state.nonZeros = append(state.nonZeros, value)
	} else {
		state.numberOfZeros += 1
	}
	return nil
}

func (state *runLengthState) flush(writer *bufio.Writer) error {
	uVarintBuffer := make([]byte, binary.MaxVarintLen16)
	n := binary.PutUvarint(uVarintBuffer, uint64(len(state.nonZeros)))
	if _, err := writer.Write(uVarintBuffer[:n]); err != nil {
		return err
	}
	n = binary.PutUvarint(uVarintBuffer, uint64(state.numberOfZeros))
	if _, err := writer.Write(uVarintBuffer[:n]); err != nil {
		return err
	}
	if err := binary.Write(writer, binary.LittleEndian, state.nonZeros); err != nil {
		return err
	}
	// state.sum += state.numberOfZeros + uint(len(state.nonZeros))
	state.numberOfZeros = 0
	state.nonZeros = state.nonZeros[:0]
	// if final {
	// 	if state.sum != uint(chunkSize) {
	// 		panic(fmt.Sprintf("sum %v should equal chunk size %v", state.sum, chunkSize))
	// 	}
	// 	state.sum = 0
	// }
	return nil
}
