package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func echo(buffer audio.IntBuffer) audio.IntBuffer { //let me guess, maybe an echo? well definitely a buffer
	bufferBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, len(buffer.Data)),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	for i := 2; i < len(buffer.Data); i += 2 {
		bufferBuffer.Data[i] = buffer.Data[i] + buffer.Data[i-2]/2
		bufferBuffer.Data[i+1] = buffer.Data[i+1] + buffer.Data[i-2]/2
	}
	return bufferBuffer
}

func multiEcho(buffer audio.IntBuffer) audio.IntBuffer { //este é meio self explaining
	dummyBuffer := buffer
	for i := 2; i < len(buffer.Data); i += 2 {
		dummyBuffer.Data[i] = dummyBuffer.Data[i] + dummyBuffer.Data[i-2]/2
		dummyBuffer.Data[i+1] = dummyBuffer.Data[i+1] + dummyBuffer.Data[i-2]/2
	}
	return dummyBuffer
}

func ampMod(buffer audio.IntBuffer) audio.IntBuffer { //amplitude modulation
	dummyBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, len(buffer.Data)),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	for i := 0; i < len(buffer.Data); i += buffer.Format.NumChannels {
		multiplier := math.Cos(float64(i))
		for c := range buffer.Format.NumChannels {
			dummyBuffer.Data[i+c] = int(float64(buffer.Data[i+c]) * multiplier)
		}
	}
	return dummyBuffer
}

func tvd(buffer audio.IntBuffer) audio.IntBuffer { //time-varying delay
	dummyBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, len(buffer.Data)+len(buffer.Data)/10),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	for i := 0; i < len(buffer.Data); i += 2 {
		dummyBuffer.Data[i+len(buffer.Data)/10] = buffer.Data[i]
	}
	return dummyBuffer
}

func main() {
	echoFlag := flag.Bool("echo", false, "enable echo generation")
	multiEchoFlag := flag.Bool("multiEcho", false, "enable multiecho generation")
	ampModFlag := flag.Bool("ampMod", false, "enable amplitude modulation generation")
	tvdFlag := flag.Bool("tvd", false, "enable time-varying delay generation")
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

	if *echoFlag {
		echoBuffer := echo(*buffer)
		file, err = os.OpenFile("echo.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
		encoder.Write(echoBuffer.AsIntBuffer())
		encoder.Close()
	}
	if *multiEchoFlag {
		multiEchoBuffer := multiEcho(*buffer)
		file, err = os.OpenFile("multiEcho.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
		encoder.Write(multiEchoBuffer.AsIntBuffer())
		encoder.Close()
	}
	if *ampModFlag {
		ampModBuffer := ampMod(*buffer)
		file, err = os.OpenFile("ampMod.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
		encoder.Write(ampModBuffer.AsIntBuffer())
		encoder.Close()
	}
	if *tvdFlag {
		tvdBuffer := tvd(*buffer)
		file, err = os.OpenFile("tvd.wav", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666)
		encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
		encoder.Write(tvdBuffer.AsIntBuffer())
		encoder.Close()
	}
}
