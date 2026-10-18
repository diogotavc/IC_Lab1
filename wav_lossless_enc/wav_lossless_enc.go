package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// func echo(buffer audio.IntBuffer, delay uint, coefficient float64) audio.IntBuffer { //let me guess, maybe an echo? well definitely a buffer
// 	bufferLen := uint(len(buffer.Data))
// 	numChannels := uint(buffer.Format.NumChannels)
// 	delay *= numChannels
// 	bufferBuffer := audio.IntBuffer{
// 		Format:         buffer.Format,
// 		Data:           make([]int, bufferLen+delay),
// 		SourceBitDepth: buffer.SourceBitDepth,
// 	}
// 	copy(bufferBuffer.Data, buffer.Data)
// 	for i := uint(0); i < bufferLen; i += numChannels {
// 		for c := range numChannels {
// 			bufferBuffer.Data[i+c+delay] += int(float64(buffer.Data[i+c]) * coefficient)
// 		}
// 	}
// 	return bufferBuffer
// }

func differentialCoding(buffer audio.IntBuffer) audio.IntBuffer { //differential coding where we store the difference between samples, reducing the number of bits and to further use the lz77 compression method :)
	bufferLen := uint(len(buffer.Data))
	numChannels := uint(buffer.Format.NumChannels)
	bufferBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	copy(bufferBuffer.Data, buffer.Data)
	for i := uint(0); i < bufferLen; i += numChannels {
		bufferBuffer.Data[i] -= buffer.Data[i+numChannels]
	}
	return bufferBuffer
}

func lz77(buffer audio.IntBuffer) map[int]int { //lz77 compression method :)
	bufferLen := uint(len(buffer.Data))
	bufferBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	copy(bufferBuffer.Data, buffer.Data)
	tmpDict := make(map[int]int)
	for _, num := range buffer.Data {
		tmpDict[num] = tmpDict[num] + 1
	}
	values := make([]int, 0)
	for _, value := range tmpDict {
		values = append(values, value)
	}
	sort.Ints(values)
	i := 0
	encoderDict := make(map[int]int)
	for number := range values {
		encoderDict[tmpDict[number]] = i
		i++
	}
	return encoderDict
}

func main() {
	differentialCodingFlag := flag.Bool("differentialCoding", false, "enable differential coding before LZ77 compression")
	// echoCoefficientFlag := flag.Float64("echoC", 0, "echo delay")
	// flag.Parse()
	file, err := os.Open(flag.Arg(0))

	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	decoder := wav.NewDecoder(file)

	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	if buffer.Format.NumChannels != 2 {
		panic("error: only stereo files supported")
	}

	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	var finalBuffer audio.IntBuffer
	if *differentialCodingFlag {
		differentialBuffer := differentialCoding(*buffer)
		finalBuffer = lz77(*differentialBuffer)
	} else {
		finalBuffer = lz77(*buffer)
	}
	file, err = os.OpenFile("thisistheaudiofileaftertheencoder.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666) //titafate iykyk
	encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
	encoder.Write(finalBuffer.AsIntBuffer())
	encoder.Close()
}
