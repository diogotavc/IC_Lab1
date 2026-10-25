package main

import (
	"container/heap"
	"flag"
	"fmt"
	"math"
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
	for i := bufferLen - 1; i >= numChannels; i-- {
		bufferBuffer.Data[i] -= buffer.Data[i-numChannels]
	}
	return bufferBuffer
}

func encoderTree(bufferData []int) *BinaryTreeNode { //lz77 compression method :)

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

func findNodePath(node *BinaryTreeNode, target int, path, pathLength uint) (uint, uint, bool) {
	if node == nil {
		return 0, 0, false
	}
	if node.isLeaf {
		return path, pathLength, node.value == target
	}

	if leftPath, leftLength, found := findNodePath(node.left, target, path<<1, pathLength+1); found {
		return leftPath, leftLength, true
	}

	if rightPath, rightLength, found := findNodePath(node.right, target, path<<1|1, pathLength+1); found {
		return rightPath, rightLength, true
	}

	return 0, 0, false
}

func huffmanMethod(bufferData []int, node *BinaryTreeNode) []byte { //helped by AI to change from bitwise to bytewise for a better storage method ([]byte)
	var result []byte
	var currentByte byte
	var bitsFilled uint

	for index, sample := range bufferData {
		path, pathLength, found := findNodePath(node, sample, 0, 0)

		if found {
			for i := int(pathLength) - 1; i >= 0; i-- {
				bit := byte((path >> i) & 1)
				currentByte = (currentByte << 1) | bit
				bitsFilled++

				if bitsFilled == 8 {
					result = append(result, currentByte)
					currentByte = 0
					bitsFilled = 0
				}
			}
		}

		if index%1000 == 0 {
			fmt.Printf("Progress: %d%%\n", index*100/len(bufferData))
		}
	}

	if bitsFilled > 0 {
		currentByte = currentByte << (8 - bitsFilled)
		result = append(result, currentByte)
	}

	return result
}

func bufferSlicing(bufferData []int) [][]int {
	var blocks [][]int
	blockSize := 4096
	blockNumber := int(math.Floor(float64(len(bufferData) / blockSize)))
	for i := 0; i <= blockNumber; i += blockSize {
		if i != blockNumber {
			blocks = append(blocks, bufferData[i:i+blockSize])
			continue
		}
		blocks = append(blocks, bufferData[i:i+(len(bufferData)-blockNumber*blockSize)])
	}
	return blocks
}

func main() {
	differentialCodingFlag := flag.Bool("differential", false, "enable differential coding before LZ77 compression")
	blockFlag := flag.Bool("slicing", false, "enable block slicing the buffer")
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

	var rootNode *BinaryTreeNode
	var encodedBytes []byte
	var blocks [][]int
	if *differentialCodingFlag {
		if *blockFlag {
			differentialBuffer := differentialCoding(*buffer)
			blocks = bufferSlicing(differentialBuffer.Data)
			for _, block := range blocks {
				rootNode = encoderTree(block)
				encodedBytes = append(encodedBytes, huffmanMethod(block, rootNode)...)
			}
			fmt.Println(encodedBytes)
			fmt.Println(len(encodedBytes))
			fmt.Println(len(buffer.Data))
		} else {
			differentialBuffer := differentialCoding(*buffer)
			rootNode = encoderTree(differentialBuffer.Data)
			encodedBytes = huffmanMethod(differentialBuffer.Data, rootNode)
			fmt.Println(encodedBytes)
			fmt.Println(len(encodedBytes))
			fmt.Println(len(buffer.Data))
		}
	} else {
		if *blockFlag {
			blocks = bufferSlicing(buffer.Data)
			for _, block := range blocks {
				rootNode = encoderTree(block)
				encodedBytes = append(encodedBytes, huffmanMethod(block, rootNode)...)
			}
			fmt.Println(encodedBytes)
			fmt.Println(len(encodedBytes))
			fmt.Println(len(buffer.Data))
		} else {
			rootNode = encoderTree(buffer.Data)
			encodedBytes = huffmanMethod(buffer.Data, rootNode)
			fmt.Println(encodedBytes)
			fmt.Println(len(encodedBytes))
			fmt.Println(len(buffer.Data))
		}
	}
	outFile, err := os.OpenFile("thisistheaudiofileaftertheencoder.bin", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666) //titafate iykyk
	if err != nil {
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

	// treeFile, err := os.OpenFile("tree.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o666) //titafate iykyk
	// if err != nil {
	// 	panic(err)
	// }

	// _, err = treeFile.Write(encodedBytes)
	// if err != nil {
	// 	treeFile.Close()
	// 	panic(err)
	// }

	// if err := treeFile.Close(); err != nil {
	// 	panic(err)
	// }
}
