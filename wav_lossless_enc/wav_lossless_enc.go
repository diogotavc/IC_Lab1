package main

import (
	"container/heap"
	"flag"
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

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

func differentialCoding(buffer audio.IntBuffer) audio.IntBuffer { //differential coding where we store the difference between samples, reducing the number of bits and to further use the lz77 compression method :)
	bufferLen := uint(len(buffer.Data))
	numChannels := uint(buffer.Format.NumChannels)
	bufferBuffer := audio.IntBuffer{
		Format:         buffer.Format,
		Data:           make([]int, bufferLen),
		SourceBitDepth: buffer.SourceBitDepth,
	}
	copy(bufferBuffer.Data, buffer.Data)
	for i := uint(1); i < bufferLen; i += numChannels {
		bufferBuffer.Data[i] -= buffer.Data[i+numChannels]
	}
	return bufferBuffer
}

func encoderTree(buffer audio.IntBuffer) *BinaryTreeNode { //lz77 compression method :)

	counter := make(map[int]uint)

	for _, num := range buffer.Data {
		counter[num]++
	}

	pq := &NodePriorityQueue{}
	heap.Init(pq)
	for num, count := range counter {
		heap.Push(pq, NodeItem{
			node: &BinaryTreeNode{
				isLeaf: true,
				value:  num,
			},
			count: count,
		})
	}
	for pq.Len() > 1 {
		leftItem := heap.Pop(pq).(NodeItem)
		rightItem := heap.Pop(pq).(NodeItem)

		parentNode := &BinaryTreeNode{
			isLeaf: false,
			left:   leftItem.node,
			right:  rightItem.node,
		}

		heap.Push(pq, NodeItem{
			node:  parentNode,
			count: leftItem.count + rightItem.count,
		})
	}

	if pq.Len() == 0 {
		return nil
	}
	return heap.Pop(pq).(NodeItem).node
}

func huffmanMethod(buffer audio.IntBuffer, node *BinaryTreeNode) uint {
	return 1
}

func main() {
	differentialCodingFlag := flag.Bool("differentialCoding", false, "enable differential coding before LZ77 compression")
	// echoCoefficientFlag := flag.Float64("echoC", 0, "echo delay")
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

	var rootNode *BinaryTreeNode
	if *differentialCodingFlag {
		differentialBuffer := differentialCoding(*buffer)
		rootNode = encoderTree(differentialBuffer)
	} else {
		rootNode = encoderTree(*buffer)
	}
	fmt.Println(huffmanMethod(*buffer, rootNode))
	//fmt.Println(encoderDict)
	// file, err = os.OpenFile("thisistheaudiofileaftertheencoder.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666) //titafate iykyk
	// encoder := wav.NewEncoder(file, buffer.Format.SampleRate, buffer.SourceBitDepth, buffer.Format.NumChannels, int(decoder.WavAudioFormat))
	// encoder.Write(finalBuffer.AsIntBuffer())
	// encoder.Close()
}
