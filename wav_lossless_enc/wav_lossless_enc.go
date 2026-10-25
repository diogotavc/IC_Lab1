package main

import (
	"container/heap"
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const blockSize = 4096

const (
	flagDifferential uint8 = 1 << 0
	flagSlicing      uint8 = 1 << 1
)

type fileHeader struct {
	NumChannels uint8
	BitDepth    uint8
	SampleRate  uint32
	NumFrames   uint64
	Flags       uint8
}

type huffmanCode struct {
	bits   uint64
	length uint
}

type symbolLength struct {
	value  int
	length uint
}

type BinaryTreeNode struct {
	// count  uint
	isLeaf bool
	value  int
	left   *BinaryTreeNode
	right  *BinaryTreeNode
}

type NodeItem struct {
	node  *BinaryTreeNode
	count uint
}

type NodePriorityQueue []*NodeItem

func (queue NodePriorityQueue) Len() int {
	return len(queue)
}

func (queue NodePriorityQueue) Less(i, j int) bool {
	return queue[i].count < queue[j].count
}

func (queue NodePriorityQueue) Swap(i, j int) {
	queue[i], queue[j] = queue[j], queue[i]
}

func (queue *NodePriorityQueue) Push(x any) {
	*queue = append(*queue, x.(*NodeItem))
}

func (queue *NodePriorityQueue) Pop() any {
	item := (*queue)[len(*queue)-1]
	*queue = (*queue)[0 : len(*queue)-1]
	return item
}

//differential coding where we store the difference between samples, reducing the number of bits
func differentialCoding(buffer audio.IntBuffer) audio.IntBuffer {
	bufferLen := uint(len(buffer.Data))
	numChannels := uint(buffer.Format.NumChannels)
	bufferBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	copy(bufferBuffer.Data, buffer.Data)
	for i := bufferLen - 1; i >= numChannels; i-- {
		bufferBuffer.Data[i] -= buffer.Data[i-numChannels]
	}
	return bufferBuffer
}

func encoderTree(bufferData []int) *BinaryTreeNode {

	counter := make(map[int]uint)

	for _, num := range bufferData {
		counter[num]++
	}

	pq := &NodePriorityQueue{}
	heap.Init(pq)
	for num, count := range counter {
		heap.Push(pq, &NodeItem{
			node: &BinaryTreeNode{
				isLeaf: true,
				value:  num,
			},
			count: count,
		})
	}
	for pq.Len() > 1 {
		leftItem := heap.Pop(pq).(*NodeItem)
		rightItem := heap.Pop(pq).(*NodeItem)

		parentNode := &BinaryTreeNode{
			isLeaf: false,
			left:   leftItem.node,
			right:  rightItem.node,
		}

		heap.Push(pq, &NodeItem{
			node:  parentNode,
			count: leftItem.count + rightItem.count,
		})
	}

	if pq.Len() == 0 {
		return nil
	}
	return heap.Pop(pq).(*NodeItem).node
}

func codeLengths(root *BinaryTreeNode) map[int]uint {
	lengths := make(map[int]uint)
	if root == nil {
		return lengths
	}
	if root.isLeaf {
		lengths[root.value] = 1
		return lengths
	}
	calculateDepths(root, 0, lengths)
	return lengths
}

func calculateDepths(node *BinaryTreeNode, depth uint, lengths map[int]uint) {
	if node.isLeaf {
		lengths[node.value] = depth
		return
	}
	calculateDepths(node.left, depth+1, lengths)
	calculateDepths(node.right, depth+1, lengths)
}

func canonicalCodes(lengths map[int]uint) (map[int]huffmanCode, []symbolLength) {
	var symbols []symbolLength
	for value, length := range lengths {
		symbols = append(symbols, symbolLength{value, length})
	}
	sort.Slice(symbols, func(i, j int) bool {
		if symbols[i].length != symbols[j].length {
			return symbols[i].length < symbols[j].length
		}
		return symbols[i].value < symbols[j].value
	})

	table := make(map[int]huffmanCode)
	var code uint64
	var prevLen uint
	for i, symbol := range symbols {
		if i > 0 {
			code++
		}
		code = code << (symbol.length - prevLen)
		table[symbol.value] = huffmanCode{code, symbol.length}
		prevLen = symbol.length
	}
	return table, symbols
}

func huffmanMethod(bufferData []int, table map[int]huffmanCode) []byte {
	var result []byte
	var currentByte byte
	var bitsFilled uint

	for _, sample := range bufferData {
		code := table[sample]
		for i := int(code.length) - 1; i >= 0; i-- {
			currentByte = currentByte<<1 | byte((code.bits>>uint(i))&1)
			bitsFilled++
			if bitsFilled == 8 {
				result = append(result, currentByte)
				currentByte = 0
				bitsFilled = 0
			}
		}
	}
	if bitsFilled > 0 {
		currentByte = currentByte << (8 - bitsFilled)
		result = append(result, currentByte)
	}
	return result
}

func encodeBlock(data []int) ([]byte, []symbolLength) {
	root := encoderTree(data)
	lengths := codeLengths(root)
	table, symbols := canonicalCodes(lengths)
	return huffmanMethod(data, table), symbols
}

func bufferSlicing(bufferData []int) [][]int {
	var blocks [][]int
	for start := 0; start < len(bufferData); start += blockSize {
		end := min(start+blockSize, len(bufferData))
		blocks = append(blocks, bufferData[start:end])
	}
	return blocks
}

func main() {
	differentialCodingFlag := flag.Bool("differential", false, "enable differential coding before LZ77 compression")
	blockFlag := flag.Bool("slicing", false, "enable block slicing the buffer")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: wav_lossless_enc [-differential] [-slicing] input.wav")
		os.Exit(1)
	}

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}
	defer file.Close()

	decoder := wav.NewDecoder(file)

	buffer, err := decoder.FullPCMBuffer()
	if err != nil {
		panic(fmt.Sprintf("error: %v", err))
	}

	if buffer.Format.NumChannels != 2 {
		panic("error: only stereo files supported")
	}

	var flags uint8
	if *differentialCodingFlag {
		flags |= flagDifferential
	}
	if *blockFlag {
		flags |= flagSlicing
	}

	numChannels := buffer.Format.NumChannels
	header := fileHeader{
		NumChannels: uint8(numChannels - 1),
		BitDepth:    uint8(buffer.SourceBitDepth - 1),
		SampleRate:  uint32(buffer.Format.SampleRate),
		NumFrames:   uint64(len(buffer.Data) / numChannels),
		Flags:       flags,
	}

	data := buffer.Data
	if *differentialCodingFlag {
		data = differentialCoding(*buffer).Data
	}

	blocks := [][]int{data}
	if *blockFlag {
		blocks = bufferSlicing(data)
	}

	var encodedBytes []byte
	for _, block := range blocks {
		bits, symbols := encodeBlock(block)
		_ = symbols	// STILL NEED TO PUT THIS INTO THE FILE TO DECODE ARGH
		encodedBytes = append(encodedBytes, bits...)
	}

	originalBytes := len(buffer.Data) * buffer.SourceBitDepth / 8
	fmt.Printf("encoded: %d bytes, original: %d bytes\n", binary.Size(header)+len(encodedBytes), originalBytes)

	outFile, err := os.OpenFile("thisistheaudiofileaftertheencoder.bin", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666) //titafate iykyk
	if err != nil {
		panic(err)
	}

	if err := binary.Write(outFile, binary.LittleEndian, header); err != nil {
		outFile.Close()
		panic(err)
	}

	_, err = outFile.Write(encodedBytes)
	if err != nil {
		outFile.Close()
		panic(err)
	}

	if err := outFile.Close(); err != nil {
		panic(err)
	}
}
