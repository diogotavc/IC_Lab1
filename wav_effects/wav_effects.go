package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func echo(buffer audio.IntBuffer, delay uint, coefficient float64) audio.IntBuffer { //let me guess, maybe an echo? well definitely a buffer
	bufferLen := uint(len(buffer.Data))
	numChannels := uint(buffer.Format.NumChannels)
	delay *= numChannels
	bufferBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, bufferLen+delay),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	copy(bufferBuffer.Data, buffer.Data)
	for i := uint(0); i < bufferLen; i += numChannels {
		for c := range numChannels {
			bufferBuffer.Data[i+c+delay] += int(float64(buffer.Data[i+c]) * coefficient)
		}
	}
	return bufferBuffer
}

func multiEcho(buffer audio.IntBuffer, delay uint, coefficient float64) audio.IntBuffer { //este é meio self explaining
	bufferLen := uint(len(buffer.Data))
	numChannels := uint(buffer.Format.NumChannels)
	delay *= numChannels
	dummyBuffer := audio.IntBuffer{
		Format: buffer.Format,
		Data: make([]int, bufferLen+delay),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	copy(dummyBuffer.Data, buffer.Data)
	for i := uint(0); i < bufferLen; i += numChannels {
		for c := range numChannels {
			dummyBuffer.Data[i+c+delay] += int(float64(dummyBuffer.Data[i+c])*coefficient)
		}
	}
	return dummyBuffer
}

func ampMod(buffer audio.IntBuffer, frequency float64) audio.IntBuffer { //amplitude modulation
	frequency *= 2 * math.Pi / float64(buffer.Format.SampleRate)
	dummyBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, len(buffer.Data)),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	for t := range len(buffer.Data) / buffer.Format.NumChannels {
		i := t * buffer.Format.NumChannels
		multiplier := math.Sin(frequency * float64(t))
		for c := range buffer.Format.NumChannels {
			dummyBuffer.Data[i+c] = int(float64(buffer.Data[i+c]) * multiplier)
		}
	}
	return dummyBuffer
}

func tvd(buffer audio.IntBuffer, delay uint) audio.IntBuffer { //time-varying delay
	channels := uint(buffer.Format.NumChannels)
	bufferLen := uint(len(buffer.Data))
	delay = delay * channels
	dummyBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, bufferLen+delay),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	for i := delay; i < bufferLen; i += channels {
		for c := range channels {
			dummyBuffer.Data[i+c+delay] = buffer.Data[i+c]
		}
	}
	return dummyBuffer
}

func main() {
	echoDelayFlag := flag.Float64("echo", 0, "enable echo generation")
	echoCoefficientFlag := flag.Float64("echoC", 0, "echo delay")
	multiEchoDelayFlag := flag.Float64("multiEcho", 0, "enable multiecho generation")
	multiEchoCoefficientFlag := flag.Float64("multiEchoC", 0, "multiple echo delay")
	ampModFlag := flag.Float64("ampMod", 0, "enable amplitude modulation generation")
	tvdFlag := flag.Int("tvd", 0, "number of samples to delay audio by")
	flag.Parse()
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

	if *echoDelayFlag > 0 {
		if *echoCoefficientFlag == 0 {
			fmt.Fprintln(os.Stderr, "echo coefficient must be given with echo delay")
		} else {
			sampleDelay := uint(*echoDelayFlag * float64(buffer.Format.SampleRate))
			echoBuffer := echo(*buffer, sampleDelay, *echoCoefficientFlag)
			file, err = os.OpenFile("echo.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
			encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
			encoder.Write(echoBuffer.AsIntBuffer())
			encoder.Close()
		}
	} else if *echoDelayFlag < 0 {
		fmt.Fprintln(os.Stderr, "echo delay must be positive")
	} else if *echoCoefficientFlag != 0 {
		fmt.Fprintln(os.Stderr, "echo coefficient should only be given with echo delay")
	}
	if *multiEchoDelayFlag > 0 {
		if *multiEchoCoefficientFlag == 0 {
			fmt.Fprintln(os.Stderr, "echo coefficient must be given with echo delay")
		} else {
			sampleDelay := uint(*multiEchoDelayFlag * float64(buffer.Format.SampleRate))
			multiEchoBuffer := multiEcho(*buffer, sampleDelay, *multiEchoCoefficientFlag)
			file, err = os.OpenFile("multiEcho.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
			encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
			encoder.Write(multiEchoBuffer.AsIntBuffer())
			encoder.Close()
		}
	} else if *multiEchoDelayFlag < 0 {
		fmt.Fprintln(os.Stderr, "multiple echo delay must be positive")
	}
	if *ampModFlag != 0 {
		ampModBuffer := ampMod(*buffer, *ampModFlag)
		file, err = os.OpenFile("ampMod.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
		encoder.Write(ampModBuffer.AsIntBuffer())
		encoder.Close()
	}
	if *tvdFlag > 0 {
		tvdBuffer := tvd(*buffer, uint(*tvdFlag))
		file, err = os.OpenFile("tvd.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
		encoder.Write(tvdBuffer.AsIntBuffer())
		encoder.Close()
	} else if *tvdFlag < 0 {
		fmt.Fprintln(os.Stderr, "delay must not be less than 0")
	}
}
