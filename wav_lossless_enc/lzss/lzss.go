package main

import "github.com/go-audio/audio"

const (
	dictionarySize = 4096 // S
	lookAheadSize  = 18   // L
	minMatchLength = 3
)

type Token struct {
	IsLiteral bool

	// IsLiteral
	Literal int

	// ¬ IsLiteral
	Pointer int
	Length  int
}

// parsing

func lzssEncode(buffer audio.IntBuffer) []Token {
	samples := buffer.Data
	tokens := make([]Token, 0, len(samples))

	position := 0
	for position < len(samples) {
		match := findLongestMatch(samples, position)

		if match.Length >= minMatchLength {
			tokens = append(tokens, match)
			position += match.Length
		} else {
			tokens = append(tokens, Token{IsLiteral: true, Literal: samples[position]})
			position++
		}
	}
	return tokens
}

func findLongestMatch(samples []int, position int) Token {
	dictionaryStart := max(0, position-dictionarySize)
	lookAheadEnd := min(len(samples), position+lookAheadSize)
	lookAhead := samples[position:lookAheadEnd]

	best := Token{}

	for candidate := position - 1; candidate >= dictionaryStart; candidate-- {
		length := 0
		for length < len(lookAhead) && samples[candidate+length] == lookAhead[length] {
			length++
		}

		if length > best.Length {
			best = Token{Pointer: position - candidate, Length: length}
			if length == len(lookAhead) {
				break
			}
		}
	}
	return best
}

func lzssDecode(tokens []Token) []int {
	samples := make([]int, 0, len(tokens))

	for _, token := range tokens {
		if token.IsLiteral {
			samples = append(samples, token.Literal)
			continue
		}

		start := len(samples) - token.Pointer
		for k := 0; k < token.Length; k++ {
			samples = append(samples, samples[start+k])
		}
	}
	return samples
}

// serialising

const (
	matchFlag   = 0
	literalFlag = 1
)

//	literalFlag, sample
//	matchFlag, pointer, length

func tokensToIntBuffer(tokens []Token, original audio.IntBuffer) audio.IntBuffer {
	symbols := make([]int, 0, 3*len(tokens))

	for _, token := range tokens {
		if token.IsLiteral {
			symbols = append(symbols, literalFlag, token.Literal)
		} else {
			symbols = append(symbols, matchFlag, token.Pointer, token.Length)
		}
	}

	return audio.IntBuffer{
		Format:         original.Format,
		Data:           symbols,
		SourceBitDepth: original.SourceBitDepth,
	}
}

func intBufferToTokens(buffer audio.IntBuffer) []Token {
	symbols := buffer.Data
	tokens := make([]Token, 0, len(symbols)/2)

	for i := 0; i < len(symbols); {
		flag := symbols[i]
		i++

		if flag == literalFlag {
			tokens = append(tokens, Token{IsLiteral: true, Literal: symbols[i]})
			i++
		} else {
			tokens = append(tokens, Token{Pointer: symbols[i], Length: symbols[i+1]})
			i += 2
		}
	}
	return tokens
}
