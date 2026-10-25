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
	var blockSize uint32
	if err := binary.Read(bufReader, binary.LittleEndian, &blockSize); err != nil {
		panic(fmt.Sprintf("error while reading block size: %v", err))
	}
	if slicing && blockSize == 0 {
		panic("slicing flag set but block size is 0")
	}
	fmt.Printf("Channels: %v\nBit depth: %v\nSampling rate: %v\nNumber of frames: %v\nDifferential: %v\nSlicing: %v\nBlock size: %v\n",
		numChannels, bitDepth, samplingRate, numFrames, differential, slicing, blockSize)

	totalSamples := numFrames * uint64(numChannels)
	samplesPerBlock := totalSamples
	if slicing {
		samplesPerBlock = uint64(blockSize)
	}

	samples := make([]int, totalSamples)
	for start := uint64(0); start < totalSamples; start += samplesPerBlock {
		blockLength := min(samplesPerBlock, totalSamples-start)

		symbols, counts, err := readTable(bufReader)
		if err != nil {
			panic(fmt.Sprintf("error while reading table of block at %v: %v", start, err))
		}
		firstCode, firstIndex := canonicalStarts(counts)

		block := samples[start : start+blockLength]
		if err := decodeBlock(bufReader, block, symbols, counts, firstCode, firstIndex); err != nil {
			panic(fmt.Sprintf("error while decoding block at %v: %v", start, err))
		}
	}

	if differential {
		for i := uint64(numChannels); i < totalSamples; i++ {
			samples[i] += samples[i-uint64(numChannels)]
		}
	}
}

func readTable(bufReader *bufio.Reader) (symbols []int, counts []uint64, err error) {
	numSymbols, err := binary.ReadUvarint(bufReader)
	if err != nil {
		return nil, nil, fmt.Errorf("error while reading number of symbols: %v", err)
	}
	longestCodeLength, err := bufReader.ReadByte()
	if err != nil {
		return nil, nil, fmt.Errorf("error while reading longest code length: %v", err)
	}

	counts = make([]uint64, uint(longestCodeLength)+1)
	var countSum uint64
	for length := uint(1); length <= uint(longestCodeLength); length++ {
		counts[length], err = binary.ReadUvarint(bufReader)
		if err != nil {
			return nil, nil, fmt.Errorf("error while reading count of length %v: %v", length, err)
		}
		countSum += counts[length]
	}
	if countSum != numSymbols {
		return nil, nil, fmt.Errorf("error: table has %v symbols but %v were counted", numSymbols, countSum)
	}

	symbols = make([]int, numSymbols)
	for i := range symbols {
		value, err := binary.ReadVarint(bufReader)
		if err != nil {
			return nil, nil, fmt.Errorf("error while reading symbol %v: %v", i, err)
		}
		symbols[i] = int(value)
	}
	return symbols, counts, nil
}

func canonicalStarts(counts []uint64) (firstCode []uint64, firstIndex []uint64) {
	firstCode = make([]uint64, len(counts))
	firstIndex = make([]uint64, len(counts))
	var code, index uint64
	for length := 1; length < len(counts); length++ {
		firstCode[length] = code
		firstIndex[length] = index
		code = (code + counts[length]) << 1
		index += counts[length]
	}
	return firstCode, firstIndex
}

type bitReader struct {
	reader   *bufio.Reader
	current  byte
	bitsLeft uint
}

func (b *bitReader) readBit() (uint64, error) {
	if b.bitsLeft == 0 {
		next, err := b.reader.ReadByte()
		if err != nil {
			return 0, err
		}
		b.current = next
		b.bitsLeft = 8
	}
	b.bitsLeft--
	return uint64(b.current>>b.bitsLeft) & 1, nil
}

func decodeBlock(bufReader *bufio.Reader, out []int, symbols []int, counts, firstCode, firstIndex []uint64) error {
	if len(out) > 0 && len(symbols) == 0 {
		return fmt.Errorf("block has %v samples but an empty table", len(out))
	}
	bits := bitReader{reader: bufReader}
	for i := range out {
		var code uint64
		for length := 1; ; length++ {
			if length >= len(counts) {
				return fmt.Errorf("sample %v: no code matches after %v bits", i, length-1)
			}
			bit, err := bits.readBit()
			if err != nil {
				return fmt.Errorf("error while reading bits of sample %v: %v", i, err)
			}
			code = code<<1 | bit
			if code >= firstCode[length] && code-firstCode[length] < counts[length] {
				out[i] = symbols[firstIndex[length]+code-firstCode[length]]
				break
			}
		}
	}
	return nil
}
