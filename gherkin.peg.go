package gherkin

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegrule uint8

const (
	ruleUnknown pegrule = iota
	ruleBegin
	ruleFeature
	ruleBackground
	ruleScenario
	ruleOutline
	ruleOutlineExamples
	ruleStep
	ruleStepArgument
	rulePyString
	rulePyStringQuote
	rulePyStringLine
	ruleTable
	ruleTableRow
	ruleTableCell
	ruleTags
	ruleTag
	ruleWord
	ruleEscapedChar
	ruleQuotedString
	ruleUntilLineEnd
	ruleLineEnd
	ruleLineComment
	ruleBlankLine
	ruleOS
	ruleWS
	ruleUntilNL
	ruleNL
	rulePegText
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25
	ruleAction26
	ruleAction27
	ruleAction28
	ruleAction29
	ruleAction30
	ruleAction31
	ruleAction32

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"Begin",
	"Feature",
	"Background",
	"Scenario",
	"Outline",
	"OutlineExamples",
	"Step",
	"StepArgument",
	"PyString",
	"PyStringQuote",
	"PyStringLine",
	"Table",
	"TableRow",
	"TableCell",
	"Tags",
	"Tag",
	"Word",
	"EscapedChar",
	"QuotedString",
	"UntilLineEnd",
	"LineEnd",
	"LineComment",
	"BlankLine",
	"OS",
	"WS",
	"UntilNL",
	"NL",
	"PegText",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",
	"Action26",
	"Action27",
	"Action28",
	"Action29",
	"Action30",
	"Action31",
	"Action32",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegrule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegrule], strconv.Quote(buffer[node.begin:node.end]))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	pegrule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.pegrule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) getToken32() token32 {
	return token32{pegrule: t.pegrule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegrule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegrule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens16) PreOrder() (<-chan state16, [][]token16) {
	s, ordered := make(chan state16, 6), t.Order()
	go func() {
		var states [8]state16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.pegrule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegrule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{pegrule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{pegrule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegrule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegrule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{pegrule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegrule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegrule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegrule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegrule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegrule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegrule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegrule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule pegrule, begin, end, depth, index int) {
	t.tree[index] = token16{pegrule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegrule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.pegrule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegrule: t.pegrule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegrule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegrule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegrule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegrule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegrule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegrule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegrule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegrule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegrule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegrule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegrule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegrule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegrule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegrule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegrule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegrule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule pegrule, begin, end, depth, index int) {
	t.tree[index] = token32{pegrule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type gherkinPeg struct {
	gherkinPegBase

	buf1    string
	buf2    string
	buftags []string
	bufcmt  string

	Buffer string
	buffer []rune
	rules  [62]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *gherkinPeg
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegrule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *gherkinPeg) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *gherkinPeg) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *gherkinPeg) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegrule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruleAction0:
			p.buf1 = buffer[begin:end]
		case ruleAction1:
			p.buf2 = buffer[begin:end]
		case ruleAction2:
			p.buf2 = p.buf2 + buffer[begin:end]
		case ruleAction3:
			p.buf2 = p.buf2 + "\n"
		case ruleAction4:
			p.beginFeature(trimWS(p.buf1), trimWSML(p.buf2), p.buftags)
			p.buftags = nil
		case ruleAction5:
			p.endFeature()
		case ruleAction6:
			p.buf1 = buffer[begin:end]
		case ruleAction7:
			p.beginBackground(trimWS(p.buf1), p.buftags)
			p.buftags = nil
		case ruleAction8:
			p.endBackground()
		case ruleAction9:
			p.buf1 = buffer[begin:end]
		case ruleAction10:
			p.beginScenario(trimWS(p.buf1), p.buftags)
			p.buftags = nil
		case ruleAction11:
			p.endScenario()
		case ruleAction12:
			p.buf1 = buffer[begin:end]
		case ruleAction13:
			p.beginOutline(trimWS(p.buf1), p.buftags)
			p.buftags = nil
		case ruleAction14:
			p.endOutline()
		case ruleAction15:
			p.buf1 = buffer[begin:end]
		case ruleAction16:
			p.beginOutlineExamples(trimWS(p.buf1))
		case ruleAction17:
			p.endOutlineExamples()
		case ruleAction18:
			p.buf1 = buffer[begin:end]
		case ruleAction19:
			p.buf2 = buffer[begin:end]
		case ruleAction20:
			p.beginStep(trimWS(p.buf1), trimWS(p.buf2))
		case ruleAction21:
			p.endStep()
		case ruleAction22:
			p.beginPyString(buffer[begin:end])
		case ruleAction23:
			p.endPyString()
		case ruleAction24:
			p.bufferPyString(buffer[begin:end])
		case ruleAction25:
			p.beginTable()
		case ruleAction26:
			p.endTable()
		case ruleAction27:
			p.beginTableRow()
		case ruleAction28:
			p.endTableRow()
		case ruleAction29:
			p.beginTableCell()
			p.endTableCell(trimWS(buffer[begin:end]))
		case ruleAction30:
			p.buftags = append(p.buftags, buffer[begin:end])
		case ruleAction31:
			p.bufcmt = buffer[begin:end]
			p.triggerComment(p.bufcmt)
		case ruleAction32:
			p.triggerBlankLine()

		}
	}
}

func (p *gherkinPeg) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegrule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Begin <- <(Feature? OS !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2, tokenIndex2, depth2 := position, tokenIndex, depth
					{
						position4 := position
						depth++
						if !_rules[ruleTags]() {
							goto l2
						}
						if buffer[position] != rune('F') {
							goto l2
						}
						position++
						if buffer[position] != rune('e') {
							goto l2
						}
						position++
						if buffer[position] != rune('a') {
							goto l2
						}
						position++
						if buffer[position] != rune('t') {
							goto l2
						}
						position++
						if buffer[position] != rune('u') {
							goto l2
						}
						position++
						if buffer[position] != rune('r') {
							goto l2
						}
						position++
						if buffer[position] != rune('e') {
							goto l2
						}
						position++
						if buffer[position] != rune(':') {
							goto l2
						}
						position++
					l5:
						{
							position6, tokenIndex6, depth6 := position, tokenIndex, depth
							if !_rules[ruleWS]() {
								goto l6
							}
							goto l5
						l6:
							position, tokenIndex, depth = position6, tokenIndex6, depth6
						}
						{
							position7 := position
							depth++
							{
								position8, tokenIndex8, depth8 := position, tokenIndex, depth
								if !_rules[ruleUntilLineEnd]() {
									goto l8
								}
								goto l9
							l8:
								position, tokenIndex, depth = position8, tokenIndex8, depth8
							}
						l9:
							depth--
							add(rulePegText, position7)
						}
						{
							add(ruleAction0, position)
						}
						{
							position11 := position
							depth++
							depth--
							add(rulePegText, position11)
						}
						{
							add(ruleAction1, position)
						}
						if !_rules[ruleLineEnd]() {
							goto l2
						}
					l13:
						{
							position14, tokenIndex14, depth14 := position, tokenIndex, depth
						l15:
							{
								position16, tokenIndex16, depth16 := position, tokenIndex, depth
								if !_rules[ruleWS]() {
									goto l16
								}
								goto l15
							l16:
								position, tokenIndex, depth = position16, tokenIndex16, depth16
							}
							{
								position17, tokenIndex17, depth17 := position, tokenIndex, depth
								{
									position18, tokenIndex18, depth18 := position, tokenIndex, depth
									if buffer[position] != rune('S') {
										goto l19
									}
									position++
									if buffer[position] != rune('c') {
										goto l19
									}
									position++
									if buffer[position] != rune('e') {
										goto l19
									}
									position++
									if buffer[position] != rune('n') {
										goto l19
									}
									position++
									if buffer[position] != rune('a') {
										goto l19
									}
									position++
									if buffer[position] != rune('r') {
										goto l19
									}
									position++
									if buffer[position] != rune('i') {
										goto l19
									}
									position++
									if buffer[position] != rune('o') {
										goto l19
									}
									position++
									if buffer[position] != rune(':') {
										goto l19
									}
									position++
									goto l18
								l19:
									position, tokenIndex, depth = position18, tokenIndex18, depth18
									{
										switch buffer[position] {
										case 'S':
											if buffer[position] != rune('S') {
												goto l17
											}
											position++
											if buffer[position] != rune('c') {
												goto l17
											}
											position++
											if buffer[position] != rune('e') {
												goto l17
											}
											position++
											if buffer[position] != rune('n') {
												goto l17
											}
											position++
											if buffer[position] != rune('a') {
												goto l17
											}
											position++
											if buffer[position] != rune('r') {
												goto l17
											}
											position++
											if buffer[position] != rune('i') {
												goto l17
											}
											position++
											if buffer[position] != rune('o') {
												goto l17
											}
											position++
											if buffer[position] != rune(' ') {
												goto l17
											}
											position++
											if buffer[position] != rune('O') {
												goto l17
											}
											position++
											if buffer[position] != rune('u') {
												goto l17
											}
											position++
											if buffer[position] != rune('t') {
												goto l17
											}
											position++
											if buffer[position] != rune('l') {
												goto l17
											}
											position++
											if buffer[position] != rune('i') {
												goto l17
											}
											position++
											if buffer[position] != rune('n') {
												goto l17
											}
											position++
											if buffer[position] != rune('e') {
												goto l17
											}
											position++
											if buffer[position] != rune(':') {
												goto l17
											}
											position++
											break
										case 'B':
											if buffer[position] != rune('B') {
												goto l17
											}
											position++
											if buffer[position] != rune('a') {
												goto l17
											}
											position++
											if buffer[position] != rune('c') {
												goto l17
											}
											position++
											if buffer[position] != rune('k') {
												goto l17
											}
											position++
											if buffer[position] != rune('g') {
												goto l17
											}
											position++
											if buffer[position] != rune('r') {
												goto l17
											}
											position++
											if buffer[position] != rune('o') {
												goto l17
											}
											position++
											if buffer[position] != rune('u') {
												goto l17
											}
											position++
											if buffer[position] != rune('n') {
												goto l17
											}
											position++
											if buffer[position] != rune('d') {
												goto l17
											}
											position++
											if buffer[position] != rune(':') {
												goto l17
											}
											position++
											break
										default:
											if buffer[position] != rune('@') {
												goto l17
											}
											position++
											if !_rules[ruleWord]() {
												goto l17
											}
											break
										}
									}

								}
							l18:
								goto l14
							l17:
								position, tokenIndex, depth = position17, tokenIndex17, depth17
							}
							{
								position21 := position
								depth++
								{
									position22, tokenIndex22, depth22 := position, tokenIndex, depth
									if !_rules[ruleUntilLineEnd]() {
										goto l22
									}
									goto l23
								l22:
									position, tokenIndex, depth = position22, tokenIndex22, depth22
								}
							l23:
								depth--
								add(rulePegText, position21)
							}
							{
								add(ruleAction2, position)
							}
							if !_rules[ruleLineEnd]() {
								goto l14
							}
							{
								add(ruleAction3, position)
							}
							goto l13
						l14:
							position, tokenIndex, depth = position14, tokenIndex14, depth14
						}
						{
							add(ruleAction4, position)
						}
					l27:
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							{
								position29, tokenIndex29, depth29 := position, tokenIndex, depth
								{
									position31 := position
									depth++
									if !_rules[ruleTags]() {
										goto l30
									}
									if buffer[position] != rune('B') {
										goto l30
									}
									position++
									if buffer[position] != rune('a') {
										goto l30
									}
									position++
									if buffer[position] != rune('c') {
										goto l30
									}
									position++
									if buffer[position] != rune('k') {
										goto l30
									}
									position++
									if buffer[position] != rune('g') {
										goto l30
									}
									position++
									if buffer[position] != rune('r') {
										goto l30
									}
									position++
									if buffer[position] != rune('o') {
										goto l30
									}
									position++
									if buffer[position] != rune('u') {
										goto l30
									}
									position++
									if buffer[position] != rune('n') {
										goto l30
									}
									position++
									if buffer[position] != rune('d') {
										goto l30
									}
									position++
									if buffer[position] != rune(':') {
										goto l30
									}
									position++
								l32:
									{
										position33, tokenIndex33, depth33 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l33
										}
										goto l32
									l33:
										position, tokenIndex, depth = position33, tokenIndex33, depth33
									}
									{
										position34 := position
										depth++
										{
											position35, tokenIndex35, depth35 := position, tokenIndex, depth
											if !_rules[ruleUntilLineEnd]() {
												goto l35
											}
											goto l36
										l35:
											position, tokenIndex, depth = position35, tokenIndex35, depth35
										}
									l36:
										depth--
										add(rulePegText, position34)
									}
									{
										add(ruleAction6, position)
									}
									if !_rules[ruleLineEnd]() {
										goto l30
									}
									{
										add(ruleAction7, position)
									}
								l39:
									{
										position40, tokenIndex40, depth40 := position, tokenIndex, depth
										{
											position41, tokenIndex41, depth41 := position, tokenIndex, depth
											if !_rules[ruleStep]() {
												goto l42
											}
											goto l41
										l42:
											position, tokenIndex, depth = position41, tokenIndex41, depth41
											if !_rules[ruleBlankLine]() {
												goto l40
											}
										}
									l41:
										goto l39
									l40:
										position, tokenIndex, depth = position40, tokenIndex40, depth40
									}
									{
										add(ruleAction8, position)
									}
									depth--
									add(ruleBackground, position31)
								}
								goto l29
							l30:
								position, tokenIndex, depth = position29, tokenIndex29, depth29
								{
									position45 := position
									depth++
									if !_rules[ruleTags]() {
										goto l44
									}
									if buffer[position] != rune('S') {
										goto l44
									}
									position++
									if buffer[position] != rune('c') {
										goto l44
									}
									position++
									if buffer[position] != rune('e') {
										goto l44
									}
									position++
									if buffer[position] != rune('n') {
										goto l44
									}
									position++
									if buffer[position] != rune('a') {
										goto l44
									}
									position++
									if buffer[position] != rune('r') {
										goto l44
									}
									position++
									if buffer[position] != rune('i') {
										goto l44
									}
									position++
									if buffer[position] != rune('o') {
										goto l44
									}
									position++
									if buffer[position] != rune(':') {
										goto l44
									}
									position++
								l46:
									{
										position47, tokenIndex47, depth47 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l47
										}
										goto l46
									l47:
										position, tokenIndex, depth = position47, tokenIndex47, depth47
									}
									{
										position48 := position
										depth++
										{
											position49, tokenIndex49, depth49 := position, tokenIndex, depth
											if !_rules[ruleUntilLineEnd]() {
												goto l49
											}
											goto l50
										l49:
											position, tokenIndex, depth = position49, tokenIndex49, depth49
										}
									l50:
										depth--
										add(rulePegText, position48)
									}
									{
										add(ruleAction9, position)
									}
									if !_rules[ruleLineEnd]() {
										goto l44
									}
									{
										add(ruleAction10, position)
									}
								l53:
									{
										position54, tokenIndex54, depth54 := position, tokenIndex, depth
										{
											position55, tokenIndex55, depth55 := position, tokenIndex, depth
											if !_rules[ruleStep]() {
												goto l56
											}
											goto l55
										l56:
											position, tokenIndex, depth = position55, tokenIndex55, depth55
											if !_rules[ruleBlankLine]() {
												goto l54
											}
										}
									l55:
										goto l53
									l54:
										position, tokenIndex, depth = position54, tokenIndex54, depth54
									}
									{
										add(ruleAction11, position)
									}
									depth--
									add(ruleScenario, position45)
								}
								goto l29
							l44:
								position, tokenIndex, depth = position29, tokenIndex29, depth29
								{
									position59 := position
									depth++
									if !_rules[ruleTags]() {
										goto l58
									}
									if buffer[position] != rune('S') {
										goto l58
									}
									position++
									if buffer[position] != rune('c') {
										goto l58
									}
									position++
									if buffer[position] != rune('e') {
										goto l58
									}
									position++
									if buffer[position] != rune('n') {
										goto l58
									}
									position++
									if buffer[position] != rune('a') {
										goto l58
									}
									position++
									if buffer[position] != rune('r') {
										goto l58
									}
									position++
									if buffer[position] != rune('i') {
										goto l58
									}
									position++
									if buffer[position] != rune('o') {
										goto l58
									}
									position++
									if buffer[position] != rune(' ') {
										goto l58
									}
									position++
									if buffer[position] != rune('O') {
										goto l58
									}
									position++
									if buffer[position] != rune('u') {
										goto l58
									}
									position++
									if buffer[position] != rune('t') {
										goto l58
									}
									position++
									if buffer[position] != rune('l') {
										goto l58
									}
									position++
									if buffer[position] != rune('i') {
										goto l58
									}
									position++
									if buffer[position] != rune('n') {
										goto l58
									}
									position++
									if buffer[position] != rune('e') {
										goto l58
									}
									position++
									if buffer[position] != rune(':') {
										goto l58
									}
									position++
								l60:
									{
										position61, tokenIndex61, depth61 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l61
										}
										goto l60
									l61:
										position, tokenIndex, depth = position61, tokenIndex61, depth61
									}
									{
										position62 := position
										depth++
										{
											position63, tokenIndex63, depth63 := position, tokenIndex, depth
											if !_rules[ruleUntilLineEnd]() {
												goto l63
											}
											goto l64
										l63:
											position, tokenIndex, depth = position63, tokenIndex63, depth63
										}
									l64:
										depth--
										add(rulePegText, position62)
									}
									{
										add(ruleAction12, position)
									}
									if !_rules[ruleLineEnd]() {
										goto l58
									}
									{
										add(ruleAction13, position)
									}
								l67:
									{
										position68, tokenIndex68, depth68 := position, tokenIndex, depth
										{
											position69, tokenIndex69, depth69 := position, tokenIndex, depth
											if !_rules[ruleStep]() {
												goto l70
											}
											goto l69
										l70:
											position, tokenIndex, depth = position69, tokenIndex69, depth69
											if !_rules[ruleBlankLine]() {
												goto l68
											}
										}
									l69:
										goto l67
									l68:
										position, tokenIndex, depth = position68, tokenIndex68, depth68
									}
								l71:
									{
										position72, tokenIndex72, depth72 := position, tokenIndex, depth
										{
											position73, tokenIndex73, depth73 := position, tokenIndex, depth
											{
												position75 := position
												depth++
												if !_rules[ruleOS]() {
													goto l74
												}
												if buffer[position] != rune('E') {
													goto l74
												}
												position++
												if buffer[position] != rune('x') {
													goto l74
												}
												position++
												if buffer[position] != rune('a') {
													goto l74
												}
												position++
												if buffer[position] != rune('m') {
													goto l74
												}
												position++
												if buffer[position] != rune('p') {
													goto l74
												}
												position++
												if buffer[position] != rune('l') {
													goto l74
												}
												position++
												if buffer[position] != rune('e') {
													goto l74
												}
												position++
												if buffer[position] != rune('s') {
													goto l74
												}
												position++
												if buffer[position] != rune(':') {
													goto l74
												}
												position++
											l76:
												{
													position77, tokenIndex77, depth77 := position, tokenIndex, depth
													if !_rules[ruleWS]() {
														goto l77
													}
													goto l76
												l77:
													position, tokenIndex, depth = position77, tokenIndex77, depth77
												}
												{
													position78 := position
													depth++
													{
														position79, tokenIndex79, depth79 := position, tokenIndex, depth
														if !_rules[ruleUntilLineEnd]() {
															goto l79
														}
														goto l80
													l79:
														position, tokenIndex, depth = position79, tokenIndex79, depth79
													}
												l80:
													depth--
													add(rulePegText, position78)
												}
												{
													add(ruleAction15, position)
												}
												if !_rules[ruleLineEnd]() {
													goto l74
												}
												{
													add(ruleAction16, position)
												}
												{
													position83, tokenIndex83, depth83 := position, tokenIndex, depth
													if !_rules[ruleTable]() {
														goto l83
													}
													goto l84
												l83:
													position, tokenIndex, depth = position83, tokenIndex83, depth83
												}
											l84:
												{
													add(ruleAction17, position)
												}
												depth--
												add(ruleOutlineExamples, position75)
											}
											goto l73
										l74:
											position, tokenIndex, depth = position73, tokenIndex73, depth73
											if !_rules[ruleBlankLine]() {
												goto l72
											}
										}
									l73:
										goto l71
									l72:
										position, tokenIndex, depth = position72, tokenIndex72, depth72
									}
									{
										add(ruleAction14, position)
									}
									depth--
									add(ruleOutline, position59)
								}
								goto l29
							l58:
								position, tokenIndex, depth = position29, tokenIndex29, depth29
								if !_rules[ruleBlankLine]() {
									goto l28
								}
							}
						l29:
							goto l27
						l28:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
						}
						{
							add(ruleAction5, position)
						}
						depth--
						add(ruleFeature, position4)
					}
					goto l3
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
			l3:
				if !_rules[ruleOS]() {
					goto l0
				}
				{
					position88, tokenIndex88, depth88 := position, tokenIndex, depth
					if !matchDot() {
						goto l88
					}
					goto l0
				l88:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
				}
				depth--
				add(ruleBegin, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Feature <- <(Tags ('F' 'e' 'a' 't' 'u' 'r' 'e' ':') WS* <UntilLineEnd?> Action0 <> Action1 LineEnd (WS* !(('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ':') / ((&('S') ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':')) | (&('B') ('B' 'a' 'c' 'k' 'g' 'r' 'o' 'u' 'n' 'd' ':')) | (&('@') ('@' Word)))) <UntilLineEnd?> Action2 LineEnd Action3)* Action4 (Background / Scenario / Outline / BlankLine)* Action5)> */
		nil,
		/* 2 Background <- <(Tags ('B' 'a' 'c' 'k' 'g' 'r' 'o' 'u' 'n' 'd' ':') WS* <UntilLineEnd?> Action6 LineEnd Action7 (Step / BlankLine)* Action8)> */
		nil,
		/* 3 Scenario <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ':') WS* <UntilLineEnd?> Action9 LineEnd Action10 (Step / BlankLine)* Action11)> */
		nil,
		/* 4 Outline <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':') WS* <UntilLineEnd?> Action12 LineEnd Action13 (Step / BlankLine)* (OutlineExamples / BlankLine)* Action14)> */
		nil,
		/* 5 OutlineExamples <- <(OS ('E' 'x' 'a' 'm' 'p' 'l' 'e' 's' ':') WS* <UntilLineEnd?> Action15 LineEnd Action16 Table? Action17)> */
		nil,
		/* 6 Step <- <(WS* <((&('B') ('B' 'u' 't')) | (&('O') ('O' 'r')) | (&('A') ('A' 'n' 'd')) | (&('T') ('T' 'h' 'e' 'n')) | (&('W') ('W' 'h' 'e' 'n')) | (&('G') ('G' 'i' 'v' 'e' 'n')))> Action18 WS* <UntilLineEnd> Action19 LineEnd Action20 StepArgument? Action21)> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
			l96:
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l97
					}
					goto l96
				l97:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
				}
				{
					position98 := position
					depth++
					{
						switch buffer[position] {
						case 'B':
							if buffer[position] != rune('B') {
								goto l94
							}
							position++
							if buffer[position] != rune('u') {
								goto l94
							}
							position++
							if buffer[position] != rune('t') {
								goto l94
							}
							position++
							break
						case 'O':
							if buffer[position] != rune('O') {
								goto l94
							}
							position++
							if buffer[position] != rune('r') {
								goto l94
							}
							position++
							break
						case 'A':
							if buffer[position] != rune('A') {
								goto l94
							}
							position++
							if buffer[position] != rune('n') {
								goto l94
							}
							position++
							if buffer[position] != rune('d') {
								goto l94
							}
							position++
							break
						case 'T':
							if buffer[position] != rune('T') {
								goto l94
							}
							position++
							if buffer[position] != rune('h') {
								goto l94
							}
							position++
							if buffer[position] != rune('e') {
								goto l94
							}
							position++
							if buffer[position] != rune('n') {
								goto l94
							}
							position++
							break
						case 'W':
							if buffer[position] != rune('W') {
								goto l94
							}
							position++
							if buffer[position] != rune('h') {
								goto l94
							}
							position++
							if buffer[position] != rune('e') {
								goto l94
							}
							position++
							if buffer[position] != rune('n') {
								goto l94
							}
							position++
							break
						default:
							if buffer[position] != rune('G') {
								goto l94
							}
							position++
							if buffer[position] != rune('i') {
								goto l94
							}
							position++
							if buffer[position] != rune('v') {
								goto l94
							}
							position++
							if buffer[position] != rune('e') {
								goto l94
							}
							position++
							if buffer[position] != rune('n') {
								goto l94
							}
							position++
							break
						}
					}

					depth--
					add(rulePegText, position98)
				}
				{
					add(ruleAction18, position)
				}
			l101:
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l102
					}
					goto l101
				l102:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
				}
				{
					position103 := position
					depth++
					if !_rules[ruleUntilLineEnd]() {
						goto l94
					}
					depth--
					add(rulePegText, position103)
				}
				{
					add(ruleAction19, position)
				}
				if !_rules[ruleLineEnd]() {
					goto l94
				}
				{
					add(ruleAction20, position)
				}
				{
					position106, tokenIndex106, depth106 := position, tokenIndex, depth
					{
						position108 := position
						depth++
						{
							position109, tokenIndex109, depth109 := position, tokenIndex, depth
							if !_rules[ruleTable]() {
								goto l110
							}
							goto l109
						l110:
							position, tokenIndex, depth = position109, tokenIndex109, depth109
							{
								position111 := position
								depth++
							l112:
								{
									position113, tokenIndex113, depth113 := position, tokenIndex, depth
								l114:
									{
										position115, tokenIndex115, depth115 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l115
										}
										goto l114
									l115:
										position, tokenIndex, depth = position115, tokenIndex115, depth115
									}
									if !_rules[ruleNL]() {
										goto l113
									}
									goto l112
								l113:
									position, tokenIndex, depth = position113, tokenIndex113, depth113
								}
								{
									position116 := position
									depth++
								l117:
									{
										position118, tokenIndex118, depth118 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l118
										}
										goto l117
									l118:
										position, tokenIndex, depth = position118, tokenIndex118, depth118
									}
									depth--
									add(rulePegText, position116)
								}
								if !_rules[rulePyStringQuote]() {
									goto l106
								}
								if !_rules[ruleNL]() {
									goto l106
								}
								{
									add(ruleAction22, position)
								}
							l120:
								{
									position121, tokenIndex121, depth121 := position, tokenIndex, depth
									{
										position122, tokenIndex122, depth122 := position, tokenIndex, depth
									l123:
										{
											position124, tokenIndex124, depth124 := position, tokenIndex, depth
											if !_rules[ruleWS]() {
												goto l124
											}
											goto l123
										l124:
											position, tokenIndex, depth = position124, tokenIndex124, depth124
										}
										if !_rules[rulePyStringQuote]() {
											goto l122
										}
										goto l121
									l122:
										position, tokenIndex, depth = position122, tokenIndex122, depth122
									}
									{
										position125 := position
										depth++
										{
											position126 := position
											depth++
											{
												position127 := position
												depth++
											l128:
												{
													position129, tokenIndex129, depth129 := position, tokenIndex, depth
													{
														position130, tokenIndex130, depth130 := position, tokenIndex, depth
														if buffer[position] != rune('\n') {
															goto l130
														}
														position++
														goto l129
													l130:
														position, tokenIndex, depth = position130, tokenIndex130, depth130
													}
													if !matchDot() {
														goto l129
													}
													goto l128
												l129:
													position, tokenIndex, depth = position129, tokenIndex129, depth129
												}
												depth--
												add(ruleUntilNL, position127)
											}
											depth--
											add(rulePegText, position126)
										}
										if !_rules[ruleNL]() {
											goto l121
										}
										{
											add(ruleAction24, position)
										}
										depth--
										add(rulePyStringLine, position125)
									}
									goto l120
								l121:
									position, tokenIndex, depth = position121, tokenIndex121, depth121
								}
							l132:
								{
									position133, tokenIndex133, depth133 := position, tokenIndex, depth
									if !_rules[ruleWS]() {
										goto l133
									}
									goto l132
								l133:
									position, tokenIndex, depth = position133, tokenIndex133, depth133
								}
								if !_rules[rulePyStringQuote]() {
									goto l106
								}
								if !_rules[ruleLineEnd]() {
									goto l106
								}
								{
									add(ruleAction23, position)
								}
								depth--
								add(rulePyString, position111)
							}
						}
					l109:
						depth--
						add(ruleStepArgument, position108)
					}
					goto l107
				l106:
					position, tokenIndex, depth = position106, tokenIndex106, depth106
				}
			l107:
				{
					add(ruleAction21, position)
				}
				depth--
				add(ruleStep, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 7 StepArgument <- <(Table / PyString)> */
		nil,
		/* 8 PyString <- <((WS* NL)* <WS*> PyStringQuote NL Action22 (!(WS* PyStringQuote) PyStringLine)* WS* PyStringQuote LineEnd Action23)> */
		nil,
		/* 9 PyStringQuote <- <('"' '"' '"')> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				if buffer[position] != rune('"') {
					goto l138
				}
				position++
				if buffer[position] != rune('"') {
					goto l138
				}
				position++
				if buffer[position] != rune('"') {
					goto l138
				}
				position++
				depth--
				add(rulePyStringQuote, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 10 PyStringLine <- <(<UntilNL> NL Action24)> */
		nil,
		/* 11 Table <- <(Action25 TableRow+ Action26)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				{
					add(ruleAction25, position)
				}
				{
					position146 := position
					depth++
					{
						add(ruleAction27, position)
					}
					if !_rules[ruleOS]() {
						goto l141
					}
					if buffer[position] != rune('|') {
						goto l141
					}
					position++
					{
						position150 := position
						depth++
						{
							position151 := position
							depth++
							{
								position154, tokenIndex154, depth154 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '|':
										if buffer[position] != rune('|') {
											goto l154
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l154
										}
										position++
										break
									default:
										if buffer[position] != rune('\r') {
											goto l154
										}
										position++
										break
									}
								}

								goto l141
							l154:
								position, tokenIndex, depth = position154, tokenIndex154, depth154
							}
							if !matchDot() {
								goto l141
							}
						l152:
							{
								position153, tokenIndex153, depth153 := position, tokenIndex, depth
								{
									position156, tokenIndex156, depth156 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l156
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l156
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l156
											}
											position++
											break
										}
									}

									goto l153
								l156:
									position, tokenIndex, depth = position156, tokenIndex156, depth156
								}
								if !matchDot() {
									goto l153
								}
								goto l152
							l153:
								position, tokenIndex, depth = position153, tokenIndex153, depth153
							}
							depth--
							add(rulePegText, position151)
						}
						if buffer[position] != rune('|') {
							goto l141
						}
						position++
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleTableCell, position150)
					}
				l148:
					{
						position149, tokenIndex149, depth149 := position, tokenIndex, depth
						{
							position159 := position
							depth++
							{
								position160 := position
								depth++
								{
									position163, tokenIndex163, depth163 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l163
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l163
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l163
											}
											position++
											break
										}
									}

									goto l149
								l163:
									position, tokenIndex, depth = position163, tokenIndex163, depth163
								}
								if !matchDot() {
									goto l149
								}
							l161:
								{
									position162, tokenIndex162, depth162 := position, tokenIndex, depth
									{
										position165, tokenIndex165, depth165 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l165
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l165
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l165
												}
												position++
												break
											}
										}

										goto l162
									l165:
										position, tokenIndex, depth = position165, tokenIndex165, depth165
									}
									if !matchDot() {
										goto l162
									}
									goto l161
								l162:
									position, tokenIndex, depth = position162, tokenIndex162, depth162
								}
								depth--
								add(rulePegText, position160)
							}
							if buffer[position] != rune('|') {
								goto l149
							}
							position++
							{
								add(ruleAction29, position)
							}
							depth--
							add(ruleTableCell, position159)
						}
						goto l148
					l149:
						position, tokenIndex, depth = position149, tokenIndex149, depth149
					}
					if !_rules[ruleLineEnd]() {
						goto l141
					}
					{
						add(ruleAction28, position)
					}
					depth--
					add(ruleTableRow, position146)
				}
			l144:
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					{
						position169 := position
						depth++
						{
							add(ruleAction27, position)
						}
						if !_rules[ruleOS]() {
							goto l145
						}
						if buffer[position] != rune('|') {
							goto l145
						}
						position++
						{
							position173 := position
							depth++
							{
								position174 := position
								depth++
								{
									position177, tokenIndex177, depth177 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l177
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l177
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l177
											}
											position++
											break
										}
									}

									goto l145
								l177:
									position, tokenIndex, depth = position177, tokenIndex177, depth177
								}
								if !matchDot() {
									goto l145
								}
							l175:
								{
									position176, tokenIndex176, depth176 := position, tokenIndex, depth
									{
										position179, tokenIndex179, depth179 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l179
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l179
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l179
												}
												position++
												break
											}
										}

										goto l176
									l179:
										position, tokenIndex, depth = position179, tokenIndex179, depth179
									}
									if !matchDot() {
										goto l176
									}
									goto l175
								l176:
									position, tokenIndex, depth = position176, tokenIndex176, depth176
								}
								depth--
								add(rulePegText, position174)
							}
							if buffer[position] != rune('|') {
								goto l145
							}
							position++
							{
								add(ruleAction29, position)
							}
							depth--
							add(ruleTableCell, position173)
						}
					l171:
						{
							position172, tokenIndex172, depth172 := position, tokenIndex, depth
							{
								position182 := position
								depth++
								{
									position183 := position
									depth++
									{
										position186, tokenIndex186, depth186 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l186
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l186
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l186
												}
												position++
												break
											}
										}

										goto l172
									l186:
										position, tokenIndex, depth = position186, tokenIndex186, depth186
									}
									if !matchDot() {
										goto l172
									}
								l184:
									{
										position185, tokenIndex185, depth185 := position, tokenIndex, depth
										{
											position188, tokenIndex188, depth188 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '|':
													if buffer[position] != rune('|') {
														goto l188
													}
													position++
													break
												case '\n':
													if buffer[position] != rune('\n') {
														goto l188
													}
													position++
													break
												default:
													if buffer[position] != rune('\r') {
														goto l188
													}
													position++
													break
												}
											}

											goto l185
										l188:
											position, tokenIndex, depth = position188, tokenIndex188, depth188
										}
										if !matchDot() {
											goto l185
										}
										goto l184
									l185:
										position, tokenIndex, depth = position185, tokenIndex185, depth185
									}
									depth--
									add(rulePegText, position183)
								}
								if buffer[position] != rune('|') {
									goto l172
								}
								position++
								{
									add(ruleAction29, position)
								}
								depth--
								add(ruleTableCell, position182)
							}
							goto l171
						l172:
							position, tokenIndex, depth = position172, tokenIndex172, depth172
						}
						if !_rules[ruleLineEnd]() {
							goto l145
						}
						{
							add(ruleAction28, position)
						}
						depth--
						add(ruleTableRow, position169)
					}
					goto l144
				l145:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
				}
				{
					add(ruleAction26, position)
				}
				depth--
				add(ruleTable, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 12 TableRow <- <(Action27 OS '|' TableCell+ LineEnd Action28)> */
		nil,
		/* 13 TableCell <- <(<(!((&('|') '|') | (&('\n') '\n') | (&('\r') '\r')) .)+> '|' Action29)> */
		nil,
		/* 14 Tags <- <((Tag+ WS* LineEnd?)* OS)> */
		func() bool {
			position195, tokenIndex195, depth195 := position, tokenIndex, depth
			{
				position196 := position
				depth++
			l197:
				{
					position198, tokenIndex198, depth198 := position, tokenIndex, depth
					{
						position201 := position
						depth++
						if !_rules[ruleOS]() {
							goto l198
						}
						if buffer[position] != rune('@') {
							goto l198
						}
						position++
						{
							position202 := position
							depth++
							if !_rules[ruleWord]() {
								goto l198
							}
							depth--
							add(rulePegText, position202)
						}
						{
							add(ruleAction30, position)
						}
						depth--
						add(ruleTag, position201)
					}
				l199:
					{
						position200, tokenIndex200, depth200 := position, tokenIndex, depth
						{
							position204 := position
							depth++
							if !_rules[ruleOS]() {
								goto l200
							}
							if buffer[position] != rune('@') {
								goto l200
							}
							position++
							{
								position205 := position
								depth++
								if !_rules[ruleWord]() {
									goto l200
								}
								depth--
								add(rulePegText, position205)
							}
							{
								add(ruleAction30, position)
							}
							depth--
							add(ruleTag, position204)
						}
						goto l199
					l200:
						position, tokenIndex, depth = position200, tokenIndex200, depth200
					}
				l207:
					{
						position208, tokenIndex208, depth208 := position, tokenIndex, depth
						if !_rules[ruleWS]() {
							goto l208
						}
						goto l207
					l208:
						position, tokenIndex, depth = position208, tokenIndex208, depth208
					}
					{
						position209, tokenIndex209, depth209 := position, tokenIndex, depth
						if !_rules[ruleLineEnd]() {
							goto l209
						}
						goto l210
					l209:
						position, tokenIndex, depth = position209, tokenIndex209, depth209
					}
				l210:
					goto l197
				l198:
					position, tokenIndex, depth = position198, tokenIndex198, depth198
				}
				if !_rules[ruleOS]() {
					goto l195
				}
				depth--
				add(ruleTags, position196)
			}
			return true
		l195:
			position, tokenIndex, depth = position195, tokenIndex195, depth195
			return false
		},
		/* 15 Tag <- <(OS '@' <Word> Action30)> */
		nil,
		/* 16 Word <- <(!((&('#') '#') | (&('"') '"') | (&(' ') ' ') | (&('\t') '\t') | (&('\n') '\n') | (&('\r') '\r')) .)+> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '#':
							if buffer[position] != rune('#') {
								goto l216
							}
							position++
							break
						case '"':
							if buffer[position] != rune('"') {
								goto l216
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l216
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l216
							}
							position++
							break
						case '\n':
							if buffer[position] != rune('\n') {
								goto l216
							}
							position++
							break
						default:
							if buffer[position] != rune('\r') {
								goto l216
							}
							position++
							break
						}
					}

					goto l212
				l216:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
				}
				if !matchDot() {
					goto l212
				}
			l214:
				{
					position215, tokenIndex215, depth215 := position, tokenIndex, depth
					{
						position218, tokenIndex218, depth218 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '#':
								if buffer[position] != rune('#') {
									goto l218
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l218
								}
								position++
								break
							case ' ':
								if buffer[position] != rune(' ') {
									goto l218
								}
								position++
								break
							case '\t':
								if buffer[position] != rune('\t') {
									goto l218
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l218
								}
								position++
								break
							default:
								if buffer[position] != rune('\r') {
									goto l218
								}
								position++
								break
							}
						}

						goto l215
					l218:
						position, tokenIndex, depth = position218, tokenIndex218, depth218
					}
					if !matchDot() {
						goto l215
					}
					goto l214
				l215:
					position, tokenIndex, depth = position215, tokenIndex215, depth215
				}
				depth--
				add(ruleWord, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 17 EscapedChar <- <('\\' .)> */
		func() bool {
			position220, tokenIndex220, depth220 := position, tokenIndex, depth
			{
				position221 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l220
				}
				position++
				if !matchDot() {
					goto l220
				}
				depth--
				add(ruleEscapedChar, position221)
			}
			return true
		l220:
			position, tokenIndex, depth = position220, tokenIndex220, depth220
			return false
		},
		/* 18 QuotedString <- <('"' (EscapedChar / (!((&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+)* '"')> */
		nil,
		/* 19 UntilLineEnd <- <(EscapedChar / (!((&('#') '#') | (&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+ / QuotedString)+> */
		func() bool {
			position223, tokenIndex223, depth223 := position, tokenIndex, depth
			{
				position224 := position
				depth++
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					if !_rules[ruleEscapedChar]() {
						goto l228
					}
					goto l227
				l228:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
					{
						position232, tokenIndex232, depth232 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '#':
								if buffer[position] != rune('#') {
									goto l232
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l232
								}
								position++
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l232
								}
								position++
								break
							default:
								if buffer[position] != rune('\n') {
									goto l232
								}
								position++
								break
							}
						}

						goto l229
					l232:
						position, tokenIndex, depth = position232, tokenIndex232, depth232
					}
					if !matchDot() {
						goto l229
					}
				l230:
					{
						position231, tokenIndex231, depth231 := position, tokenIndex, depth
						{
							position234, tokenIndex234, depth234 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l234
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l234
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l234
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l234
									}
									position++
									break
								}
							}

							goto l231
						l234:
							position, tokenIndex, depth = position234, tokenIndex234, depth234
						}
						if !matchDot() {
							goto l231
						}
						goto l230
					l231:
						position, tokenIndex, depth = position231, tokenIndex231, depth231
					}
					goto l227
				l229:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
					{
						position236 := position
						depth++
						if buffer[position] != rune('"') {
							goto l223
						}
						position++
					l237:
						{
							position238, tokenIndex238, depth238 := position, tokenIndex, depth
							{
								position239, tokenIndex239, depth239 := position, tokenIndex, depth
								if !_rules[ruleEscapedChar]() {
									goto l240
								}
								goto l239
							l240:
								position, tokenIndex, depth = position239, tokenIndex239, depth239
								{
									position243, tokenIndex243, depth243 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '"':
											if buffer[position] != rune('"') {
												goto l243
											}
											position++
											break
										case '\\':
											if buffer[position] != rune('\\') {
												goto l243
											}
											position++
											break
										default:
											if buffer[position] != rune('\n') {
												goto l243
											}
											position++
											break
										}
									}

									goto l238
								l243:
									position, tokenIndex, depth = position243, tokenIndex243, depth243
								}
								if !matchDot() {
									goto l238
								}
							l241:
								{
									position242, tokenIndex242, depth242 := position, tokenIndex, depth
									{
										position245, tokenIndex245, depth245 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l245
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l245
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l245
												}
												position++
												break
											}
										}

										goto l242
									l245:
										position, tokenIndex, depth = position245, tokenIndex245, depth245
									}
									if !matchDot() {
										goto l242
									}
									goto l241
								l242:
									position, tokenIndex, depth = position242, tokenIndex242, depth242
								}
							}
						l239:
							goto l237
						l238:
							position, tokenIndex, depth = position238, tokenIndex238, depth238
						}
						if buffer[position] != rune('"') {
							goto l223
						}
						position++
						depth--
						add(ruleQuotedString, position236)
					}
				}
			l227:
			l225:
				{
					position226, tokenIndex226, depth226 := position, tokenIndex, depth
					{
						position247, tokenIndex247, depth247 := position, tokenIndex, depth
						if !_rules[ruleEscapedChar]() {
							goto l248
						}
						goto l247
					l248:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
						{
							position252, tokenIndex252, depth252 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l252
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l252
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l252
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l252
									}
									position++
									break
								}
							}

							goto l249
						l252:
							position, tokenIndex, depth = position252, tokenIndex252, depth252
						}
						if !matchDot() {
							goto l249
						}
					l250:
						{
							position251, tokenIndex251, depth251 := position, tokenIndex, depth
							{
								position254, tokenIndex254, depth254 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '#':
										if buffer[position] != rune('#') {
											goto l254
										}
										position++
										break
									case '"':
										if buffer[position] != rune('"') {
											goto l254
										}
										position++
										break
									case '\\':
										if buffer[position] != rune('\\') {
											goto l254
										}
										position++
										break
									default:
										if buffer[position] != rune('\n') {
											goto l254
										}
										position++
										break
									}
								}

								goto l251
							l254:
								position, tokenIndex, depth = position254, tokenIndex254, depth254
							}
							if !matchDot() {
								goto l251
							}
							goto l250
						l251:
							position, tokenIndex, depth = position251, tokenIndex251, depth251
						}
						goto l247
					l249:
						position, tokenIndex, depth = position247, tokenIndex247, depth247
						{
							position256 := position
							depth++
							if buffer[position] != rune('"') {
								goto l226
							}
							position++
						l257:
							{
								position258, tokenIndex258, depth258 := position, tokenIndex, depth
								{
									position259, tokenIndex259, depth259 := position, tokenIndex, depth
									if !_rules[ruleEscapedChar]() {
										goto l260
									}
									goto l259
								l260:
									position, tokenIndex, depth = position259, tokenIndex259, depth259
									{
										position263, tokenIndex263, depth263 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l263
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l263
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l263
												}
												position++
												break
											}
										}

										goto l258
									l263:
										position, tokenIndex, depth = position263, tokenIndex263, depth263
									}
									if !matchDot() {
										goto l258
									}
								l261:
									{
										position262, tokenIndex262, depth262 := position, tokenIndex, depth
										{
											position265, tokenIndex265, depth265 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '"':
													if buffer[position] != rune('"') {
														goto l265
													}
													position++
													break
												case '\\':
													if buffer[position] != rune('\\') {
														goto l265
													}
													position++
													break
												default:
													if buffer[position] != rune('\n') {
														goto l265
													}
													position++
													break
												}
											}

											goto l262
										l265:
											position, tokenIndex, depth = position265, tokenIndex265, depth265
										}
										if !matchDot() {
											goto l262
										}
										goto l261
									l262:
										position, tokenIndex, depth = position262, tokenIndex262, depth262
									}
								}
							l259:
								goto l257
							l258:
								position, tokenIndex, depth = position258, tokenIndex258, depth258
							}
							if buffer[position] != rune('"') {
								goto l226
							}
							position++
							depth--
							add(ruleQuotedString, position256)
						}
					}
				l247:
					goto l225
				l226:
					position, tokenIndex, depth = position226, tokenIndex226, depth226
				}
				depth--
				add(ruleUntilLineEnd, position224)
			}
			return true
		l223:
			position, tokenIndex, depth = position223, tokenIndex223, depth223
			return false
		},
		/* 20 LineEnd <- <(WS* LineComment? NL)> */
		func() bool {
			position267, tokenIndex267, depth267 := position, tokenIndex, depth
			{
				position268 := position
				depth++
			l269:
				{
					position270, tokenIndex270, depth270 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l270
					}
					goto l269
				l270:
					position, tokenIndex, depth = position270, tokenIndex270, depth270
				}
				{
					position271, tokenIndex271, depth271 := position, tokenIndex, depth
					if !_rules[ruleLineComment]() {
						goto l271
					}
					goto l272
				l271:
					position, tokenIndex, depth = position271, tokenIndex271, depth271
				}
			l272:
				if !_rules[ruleNL]() {
					goto l267
				}
				depth--
				add(ruleLineEnd, position268)
			}
			return true
		l267:
			position, tokenIndex, depth = position267, tokenIndex267, depth267
			return false
		},
		/* 21 LineComment <- <('#' <(!'\n' .)*> Action31)> */
		func() bool {
			position273, tokenIndex273, depth273 := position, tokenIndex, depth
			{
				position274 := position
				depth++
				if buffer[position] != rune('#') {
					goto l273
				}
				position++
				{
					position275 := position
					depth++
				l276:
					{
						position277, tokenIndex277, depth277 := position, tokenIndex, depth
						{
							position278, tokenIndex278, depth278 := position, tokenIndex, depth
							if buffer[position] != rune('\n') {
								goto l278
							}
							position++
							goto l277
						l278:
							position, tokenIndex, depth = position278, tokenIndex278, depth278
						}
						if !matchDot() {
							goto l277
						}
						goto l276
					l277:
						position, tokenIndex, depth = position277, tokenIndex277, depth277
					}
					depth--
					add(rulePegText, position275)
				}
				{
					add(ruleAction31, position)
				}
				depth--
				add(ruleLineComment, position274)
			}
			return true
		l273:
			position, tokenIndex, depth = position273, tokenIndex273, depth273
			return false
		},
		/* 22 BlankLine <- <(((WS LineEnd) / (LineComment? NL)) Action32)> */
		func() bool {
			position280, tokenIndex280, depth280 := position, tokenIndex, depth
			{
				position281 := position
				depth++
				{
					position282, tokenIndex282, depth282 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l283
					}
					if !_rules[ruleLineEnd]() {
						goto l283
					}
					goto l282
				l283:
					position, tokenIndex, depth = position282, tokenIndex282, depth282
					{
						position284, tokenIndex284, depth284 := position, tokenIndex, depth
						if !_rules[ruleLineComment]() {
							goto l284
						}
						goto l285
					l284:
						position, tokenIndex, depth = position284, tokenIndex284, depth284
					}
				l285:
					if !_rules[ruleNL]() {
						goto l280
					}
				}
			l282:
				{
					add(ruleAction32, position)
				}
				depth--
				add(ruleBlankLine, position281)
			}
			return true
		l280:
			position, tokenIndex, depth = position280, tokenIndex280, depth280
			return false
		},
		/* 23 OS <- <(NL / WS)*> */
		func() bool {
			{
				position288 := position
				depth++
			l289:
				{
					position290, tokenIndex290, depth290 := position, tokenIndex, depth
					{
						position291, tokenIndex291, depth291 := position, tokenIndex, depth
						if !_rules[ruleNL]() {
							goto l292
						}
						goto l291
					l292:
						position, tokenIndex, depth = position291, tokenIndex291, depth291
						if !_rules[ruleWS]() {
							goto l290
						}
					}
				l291:
					goto l289
				l290:
					position, tokenIndex, depth = position290, tokenIndex290, depth290
				}
				depth--
				add(ruleOS, position288)
			}
			return true
		},
		/* 24 WS <- <(' ' / '\t')> */
		func() bool {
			position293, tokenIndex293, depth293 := position, tokenIndex, depth
			{
				position294 := position
				depth++
				{
					position295, tokenIndex295, depth295 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l296
					}
					position++
					goto l295
				l296:
					position, tokenIndex, depth = position295, tokenIndex295, depth295
					if buffer[position] != rune('\t') {
						goto l293
					}
					position++
				}
			l295:
				depth--
				add(ruleWS, position294)
			}
			return true
		l293:
			position, tokenIndex, depth = position293, tokenIndex293, depth293
			return false
		},
		/* 25 UntilNL <- <(!'\n' .)*> */
		nil,
		/* 26 NL <- <('\n' / '\r' / ('\r' '\n'))> */
		func() bool {
			position298, tokenIndex298, depth298 := position, tokenIndex, depth
			{
				position299 := position
				depth++
				{
					position300, tokenIndex300, depth300 := position, tokenIndex, depth
					if buffer[position] != rune('\n') {
						goto l301
					}
					position++
					goto l300
				l301:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
					if buffer[position] != rune('\r') {
						goto l302
					}
					position++
					goto l300
				l302:
					position, tokenIndex, depth = position300, tokenIndex300, depth300
					if buffer[position] != rune('\r') {
						goto l298
					}
					position++
					if buffer[position] != rune('\n') {
						goto l298
					}
					position++
				}
			l300:
				depth--
				add(ruleNL, position299)
			}
			return true
		l298:
			position, tokenIndex, depth = position298, tokenIndex298, depth298
			return false
		},
		nil,
		/* 29 Action0 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 30 Action1 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 31 Action2 <- <{ p.buf2 = p.buf2 + buffer[begin:end] }> */
		nil,
		/* 32 Action3 <- <{ p.buf2 = p.buf2 + "\n" }> */
		nil,
		/* 33 Action4 <- <{ p.beginFeature(trimWS(p.buf1), trimWSML(p.buf2), p.buftags); p.buftags = nil }> */
		nil,
		/* 34 Action5 <- <{ p.endFeature() }> */
		nil,
		/* 35 Action6 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 36 Action7 <- <{ p.beginBackground(trimWS(p.buf1), p.buftags); p.buftags = nil }> */
		nil,
		/* 37 Action8 <- <{ p.endBackground() }> */
		nil,
		/* 38 Action9 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 39 Action10 <- <{ p.beginScenario(trimWS(p.buf1), p.buftags); p.buftags = nil }> */
		nil,
		/* 40 Action11 <- <{ p.endScenario() }> */
		nil,
		/* 41 Action12 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 42 Action13 <- <{ p.beginOutline(trimWS(p.buf1), p.buftags); p.buftags = nil }> */
		nil,
		/* 43 Action14 <- <{ p.endOutline() }> */
		nil,
		/* 44 Action15 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 45 Action16 <- <{ p.beginOutlineExamples(trimWS(p.buf1)) }> */
		nil,
		/* 46 Action17 <- <{ p.endOutlineExamples() }> */
		nil,
		/* 47 Action18 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 48 Action19 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 49 Action20 <- <{ p.beginStep(trimWS(p.buf1), trimWS(p.buf2)) }> */
		nil,
		/* 50 Action21 <- <{ p.endStep() }> */
		nil,
		/* 51 Action22 <- <{ p.beginPyString(buffer[begin:end]) }> */
		nil,
		/* 52 Action23 <- <{ p.endPyString() }> */
		nil,
		/* 53 Action24 <- <{ p.bufferPyString(buffer[begin:end]) }> */
		nil,
		/* 54 Action25 <- <{ p.beginTable() }> */
		nil,
		/* 55 Action26 <- <{ p.endTable() }> */
		nil,
		/* 56 Action27 <- <{ p.beginTableRow() }> */
		nil,
		/* 57 Action28 <- <{ p.endTableRow() }> */
		nil,
		/* 58 Action29 <- <{ p.beginTableCell(); p.endTableCell(trimWS(buffer[begin:end])) }> */
		nil,
		/* 59 Action30 <- <{ p.buftags = append(p.buftags, buffer[begin:end]) }> */
		nil,
		/* 60 Action31 <- <{ p.bufcmt = buffer[begin:end]; p.triggerComment(p.bufcmt) }> */
		nil,
		/* 61 Action32 <- <{ p.triggerBlankLine() }> */
		nil,
	}
	p.rules = _rules
}
