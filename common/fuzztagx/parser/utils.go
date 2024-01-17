package parser

import (
	"golang.org/x/exp/maps"
	"strings"
)

type trieNode struct {
	children   map[rune]*trieNode
	failure    *trieNode
	patternLen int
	id         int
	flag       int // The mark of the node can be used to mark the end node
}

// IndexAllSubstrings Only traverse once to find all substring positions
// The return value is a two-dimensional array, each element is a [2]int type Matching result, the first element is the rule index, the second element is the index position
func IndexAllSubstrings(s string, patterns ...string) (result [][2]int) {
	// Build a trie tree
	root := &trieNode{
		children:   make(map[rune]*trieNode),
		failure:    nil,
		flag:       0,
		patternLen: 0,
	}

	for patternIndex, pattern := range patterns {
		node := root
		for _, char := range pattern {
			if _, ok := node.children[char]; !ok {
				node.children[char] = &trieNode{
					children:   make(map[rune]*trieNode),
					failure:    nil,
					flag:       0,
					patternLen: 0,
					id:         patternIndex,
				}
			}
			node = node.children[char]
		}
		node.flag = 1
		node.patternLen = len(pattern)
	}
	// Build Failure
	queue := make([]*trieNode, 0)
	root.failure = root

	for _, child := range root.children {
		child.failure = root
		queue = append(queue, child)
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		for char, child := range node.children {
			queue = append(queue, child)
			failure := node.failure

			for failure != root && failure.children[char] == nil {
				failure = failure.failure
			}
			if next := failure.children[char]; next != nil {
				child.failure = next
				child.flag = child.flag | next.flag
			} else {
				child.failure = root
			}
		}
	}

	// Find
	node := root
	for i, char := range s {
		for node != root && node.children[char] == nil {
			node = node.failure
		}

		if next := node.children[char]; next != nil {
			node = next
			if node.flag == 1 {
				result = append(result, [2]int{node.id, i - node.patternLen + 1})
			}
		}
	}
	return
}

type Escaper struct {
	escapeSymbol string
	escapeChars  map[string]string
}

func (e *Escaper) Escape(s string) string {
	keys := maps.Keys(e.escapeChars)
	poses := IndexAllSubstrings(s, keys...)
	res := ""
	pre := 0
	for _, pos := range poses {
		key := keys[pos[0]]
		res += s[pre:pos[1]]
		res += (e.escapeSymbol + key)
		pre = pos[1] + len(key)
	}
	res += s[pre:]
	return res
}
func (e *Escaper) Unescape(s string) (string, error) {
	// Build a trie tree
	root := &trieNode{
		children:   make(map[rune]*trieNode),
		failure:    nil,
		flag:       0,
		patternLen: 0,
	}
	patterns := []string{}
	for pattern, _ := range e.escapeChars {
		patterns = append(patterns, pattern)
		node := root
		for _, char := range pattern {
			if _, ok := node.children[char]; !ok {
				node.children[char] = &trieNode{
					children:   make(map[rune]*trieNode),
					failure:    nil,
					flag:       0,
					patternLen: 0,
					id:         len(patterns) - 1,
				}
			}
			node = node.children[char]
		}
		node.flag = 1
		node.patternLen = len(pattern)
	}

	var result string
	escapeState := false
	node := root
	data := s
	for {
		if escapeState {
			escapeState = false
			runeData := []rune(data)
			for i := 0; i < len(runeData); i++ {
				ch := runeData[i]
				if node.children[ch] != nil {
					node = node.children[ch]
					if node.flag == 1 { // Match success
						result += patterns[node.id]
						data = string(runeData[i+1:])
						node = root
						break
					}
				} else {
					result += string(runeData[:i])
					data = string(runeData[i:])
					node = root
					break
				}
			}
		} else {
			index := strings.Index(data, e.escapeSymbol) // Find the first escape character after
			if index != -1 {
				result += data[:index]
				data = data[index+len(e.escapeSymbol):]
				escapeState = true
			} else {
				result += data
				break
			}
		}
	}
	return result, nil
}
func NewEscaper(escapeSymbol string, charsMap map[string]string) *Escaper {
	if _, ok := charsMap[escapeSymbol]; !ok {
		charsMap[escapeSymbol] = escapeSymbol
	}
	return &Escaper{
		escapeSymbol: escapeSymbol,
		escapeChars:  charsMap,
	}
}
func NewDefaultEscaper(chars ...string) *Escaper {
	m := map[string]string{}
	for _, char := range chars {
		m[char] = char
	}
	return NewEscaper(`\`, m)
}
