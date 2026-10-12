package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-audio/wav"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	bin_exp, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	bin_mask := 1 << bin_exp - 1

	decoder := wav.NewDecoder(file)

	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	if buffer.Format.NumChannels != 2 {
		panic("error: only stereo files supported")
	}

	for i := 0; i < len(buffer.Data); i += 2 {
		buffer.Data[i] = buffer.Data[i] &^ bin_mask
		buffer.Data[i+1] = buffer.Data[i+1] &^ bin_mask
	}
	file, err = os.OpenFile("../example_w_bin.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)

	encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
	encoder.Write(buffer.AsIntBuffer())
	encoder.Close()
}
