package alogrithm

import "fmt"

// node
type Node interface {
	Next() []Node
	Prev() []Node
	Handler(*Node) any
}

type TrSCC struct {
	dfn       map[Node]int
	low       map[Node]int
	id        map[Node]int
	stack     []Node
	in_stack  map[Node]bool
	in_scc    map[Node]bool
	edges     map[Edge]struct{}
	timestamp int
	stack_cap int
	scc_cnt   int
	top       int

	result []SccResultItem
}

func NewTrSCC() *TrSCC {
	return &TrSCC{
		dfn:       make(map[Node]int),
		low:       make(map[Node]int),
		id:        make(map[Node]int),
		stack:     make([]Node, 50),
		in_stack:  make(map[Node]bool),
		in_scc:    make(map[Node]bool),
		edges:     make(map[Edge]struct{}),
		result:    make([]SccResultItem, 0),
		stack_cap: 50,
		timestamp: 1,
		scc_cnt:   0,
		top:       0,
	}
}

type Edge struct {
	from Node
	to   Node
}

func NewEdge(from Node, to Node) *Edge {
	return &Edge{
		from: from,
		to:   to,
	}
}

type SccResult []SccResultItem

func (scc SccResult) GetScc(n Node) SccResultItem {
	for _, item := range scc {
		if item.InNodes(n) {
			return item
		}
	}
	return SccResultItem{}
}

type SccResultItem struct {
	nodes  map[Node]struct{}
	input  map[Node]struct{}
	output map[Node]struct{}
}

func (scc *SccResultItem) InNodes(n Node) bool {
	_, ok := scc.nodes[n]
	return ok
}

func (scc *SccResultItem) InInput(n Node) bool {
	_, ok := scc.input[n]
	return ok
}

func (scc *SccResultItem) InOutput(n Node) bool {
	_, ok := scc.output[n]
	return ok
}

func NewSccResult(scc *TrSCC) *SccResultItem {
	return &SccResultItem{
		nodes:  make(map[Node]struct{}),
		input:  make(map[Node]struct{}),
		output: make(map[Node]struct{}),
	}
}

func Run(rootNode Node) []SccResultItem {
	scc := NewTrSCC()
	scc.Tarjan(rootNode)
	scc.finish()
	return scc.result
}

func (scc *TrSCC) finish() {
	// . Add the in-degree and out-degree
	for k, _ := range scc.edges {
		id_f, ok1 := scc.id[k.from]
		id_t, ok2 := scc.id[k.to]
		if !ok1 || !ok2 {
			fmt.Println("node not in scc")
		} else if id_f != id_t {
			scc.result[id_f-1].output[k.to] = struct{}{}
			scc.result[id_t-1].input[k.from] = struct{}{}
		}
	}
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// . Calculate the strongly connected component
func (scc *TrSCC) Tarjan(node Node) {
	dfn := scc.dfn
	low := scc.low
	in_stack := scc.in_stack
	in_scc := scc.in_scc
	id := scc.id
	// to see if the stack will overflow
	if scc.timestamp == scc.stack_cap {
		scc.stack_cap *= 2
		stack_new := make([]Node, scc.stack_cap)
		copy(stack_new, scc.stack)
		scc.stack = stack_new
	}
	stack := scc.stack

	dfn[node] = scc.timestamp
	low[node] = scc.timestamp
	scc.timestamp += 1
	stack[scc.top] = node
	scc.top += 1
	in_stack[node] = true
	for _, n := range node.Next() {
		// . Add edge
		edge := NewEdge(node, n)
		scc.edges[*edge] = struct{}{}
		if dfn[n] == 0 {
			scc.Tarjan(n)
			low[node] = Min(low[node], low[n])
		} else if in_stack[n] {
			low[node] = Min(low[node], dfn[n])
		}
	}

	if dfn[node] == low[node] {
		// . Add a strongly connected component
		sccResult := NewSccResult(scc)
		sccResult.nodes[node] = struct{}{}
		// In the strongly connected component,
		in_scc[node] = true
		// . The number of sccs + 1
		scc.scc_cnt += 1
		id[node] = scc.scc_cnt
		// fmt.Println("id:", id[node])
		// . The last node in scc that is pushed onto the stack is used as
		var y Node
		y = stack[scc.top-1]
		for y != node {
			// In the strongly connected component,
			in_scc[y] = true
			sccResult.nodes[y] = struct{}{}
			// pops the stack
			scc.top -= 1
			// tag
			in_stack[y] = false
			id[y] = scc.scc_cnt
			if scc.top < 1 {
				break
			}
			y = stack[scc.top-1]
		}
		scc.result = append(scc.result, *sccResult)
		scc.top -= 1
	}
}
