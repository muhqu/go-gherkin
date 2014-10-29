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
	rules  [61]func() bool
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
			p.beginOutlineExamples()
		case ruleAction16:
			p.endOutlineExamples()
		case ruleAction17:
			p.buf1 = buffer[begin:end]
		case ruleAction18:
			p.buf2 = buffer[begin:end]
		case ruleAction19:
			p.beginStep(trimWS(p.buf1), trimWS(p.buf2))
		case ruleAction20:
			p.endStep()
		case ruleAction21:
			p.beginPyString(buffer[begin:end])
		case ruleAction22:
			p.endPyString()
		case ruleAction23:
			p.bufferPyString(buffer[begin:end])
		case ruleAction24:
			p.beginTable()
		case ruleAction25:
			p.endTable()
		case ruleAction26:
			p.beginTableRow()
		case ruleAction27:
			p.endTableRow()
		case ruleAction28:
			p.beginTableCell()
			p.endTableCell(trimWS(buffer[begin:end]))
		case ruleAction29:
			p.buftags = append(p.buftags, buffer[begin:end])
		case ruleAction30:
			p.bufcmt = buffer[begin:end]
			p.triggerComment(p.bufcmt)
		case ruleAction31:
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
									{
										position71, tokenIndex71, depth71 := position, tokenIndex, depth
										{
											position73 := position
											depth++
											if !_rules[ruleOS]() {
												goto l71
											}
											if buffer[position] != rune('E') {
												goto l71
											}
											position++
											if buffer[position] != rune('x') {
												goto l71
											}
											position++
											if buffer[position] != rune('a') {
												goto l71
											}
											position++
											if buffer[position] != rune('m') {
												goto l71
											}
											position++
											if buffer[position] != rune('p') {
												goto l71
											}
											position++
											if buffer[position] != rune('l') {
												goto l71
											}
											position++
											if buffer[position] != rune('e') {
												goto l71
											}
											position++
											if buffer[position] != rune('s') {
												goto l71
											}
											position++
											if buffer[position] != rune(':') {
												goto l71
											}
											position++
											if !_rules[ruleLineEnd]() {
												goto l71
											}
											{
												add(ruleAction15, position)
											}
											{
												position75, tokenIndex75, depth75 := position, tokenIndex, depth
												if !_rules[ruleTable]() {
													goto l75
												}
												goto l76
											l75:
												position, tokenIndex, depth = position75, tokenIndex75, depth75
											}
										l76:
											{
												add(ruleAction16, position)
											}
											depth--
											add(ruleOutlineExamples, position73)
										}
										goto l72
									l71:
										position, tokenIndex, depth = position71, tokenIndex71, depth71
									}
								l72:
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
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if !matchDot() {
						goto l80
					}
					goto l0
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
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
		/* 4 Outline <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':') WS* <UntilLineEnd?> Action12 LineEnd Action13 (Step / BlankLine)* OutlineExamples? Action14)> */
		nil,
		/* 5 OutlineExamples <- <(OS ('E' 'x' 'a' 'm' 'p' 'l' 'e' 's' ':') LineEnd Action15 Table? Action16)> */
		nil,
		/* 6 Step <- <(WS* <((&('B') ('B' 'u' 't')) | (&('O') ('O' 'r')) | (&('A') ('A' 'n' 'd')) | (&('T') ('T' 'h' 'e' 'n')) | (&('W') ('W' 'h' 'e' 'n')) | (&('G') ('G' 'i' 'v' 'e' 'n')))> Action17 WS* <UntilLineEnd> Action18 LineEnd Action19 StepArgument? Action20)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
			l88:
				{
					position89, tokenIndex89, depth89 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l89
					}
					goto l88
				l89:
					position, tokenIndex, depth = position89, tokenIndex89, depth89
				}
				{
					position90 := position
					depth++
					{
						switch buffer[position] {
						case 'B':
							if buffer[position] != rune('B') {
								goto l86
							}
							position++
							if buffer[position] != rune('u') {
								goto l86
							}
							position++
							if buffer[position] != rune('t') {
								goto l86
							}
							position++
							break
						case 'O':
							if buffer[position] != rune('O') {
								goto l86
							}
							position++
							if buffer[position] != rune('r') {
								goto l86
							}
							position++
							break
						case 'A':
							if buffer[position] != rune('A') {
								goto l86
							}
							position++
							if buffer[position] != rune('n') {
								goto l86
							}
							position++
							if buffer[position] != rune('d') {
								goto l86
							}
							position++
							break
						case 'T':
							if buffer[position] != rune('T') {
								goto l86
							}
							position++
							if buffer[position] != rune('h') {
								goto l86
							}
							position++
							if buffer[position] != rune('e') {
								goto l86
							}
							position++
							if buffer[position] != rune('n') {
								goto l86
							}
							position++
							break
						case 'W':
							if buffer[position] != rune('W') {
								goto l86
							}
							position++
							if buffer[position] != rune('h') {
								goto l86
							}
							position++
							if buffer[position] != rune('e') {
								goto l86
							}
							position++
							if buffer[position] != rune('n') {
								goto l86
							}
							position++
							break
						default:
							if buffer[position] != rune('G') {
								goto l86
							}
							position++
							if buffer[position] != rune('i') {
								goto l86
							}
							position++
							if buffer[position] != rune('v') {
								goto l86
							}
							position++
							if buffer[position] != rune('e') {
								goto l86
							}
							position++
							if buffer[position] != rune('n') {
								goto l86
							}
							position++
							break
						}
					}

					depth--
					add(rulePegText, position90)
				}
				{
					add(ruleAction17, position)
				}
			l93:
				{
					position94, tokenIndex94, depth94 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l94
					}
					goto l93
				l94:
					position, tokenIndex, depth = position94, tokenIndex94, depth94
				}
				{
					position95 := position
					depth++
					if !_rules[ruleUntilLineEnd]() {
						goto l86
					}
					depth--
					add(rulePegText, position95)
				}
				{
					add(ruleAction18, position)
				}
				if !_rules[ruleLineEnd]() {
					goto l86
				}
				{
					add(ruleAction19, position)
				}
				{
					position98, tokenIndex98, depth98 := position, tokenIndex, depth
					{
						position100 := position
						depth++
						{
							position101, tokenIndex101, depth101 := position, tokenIndex, depth
							if !_rules[ruleTable]() {
								goto l102
							}
							goto l101
						l102:
							position, tokenIndex, depth = position101, tokenIndex101, depth101
							{
								position103 := position
								depth++
							l104:
								{
									position105, tokenIndex105, depth105 := position, tokenIndex, depth
								l106:
									{
										position107, tokenIndex107, depth107 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l107
										}
										goto l106
									l107:
										position, tokenIndex, depth = position107, tokenIndex107, depth107
									}
									if !_rules[ruleNL]() {
										goto l105
									}
									goto l104
								l105:
									position, tokenIndex, depth = position105, tokenIndex105, depth105
								}
								{
									position108 := position
									depth++
								l109:
									{
										position110, tokenIndex110, depth110 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l110
										}
										goto l109
									l110:
										position, tokenIndex, depth = position110, tokenIndex110, depth110
									}
									depth--
									add(rulePegText, position108)
								}
								if !_rules[rulePyStringQuote]() {
									goto l98
								}
								if !_rules[ruleNL]() {
									goto l98
								}
								{
									add(ruleAction21, position)
								}
							l112:
								{
									position113, tokenIndex113, depth113 := position, tokenIndex, depth
									{
										position114, tokenIndex114, depth114 := position, tokenIndex, depth
									l115:
										{
											position116, tokenIndex116, depth116 := position, tokenIndex, depth
											if !_rules[ruleWS]() {
												goto l116
											}
											goto l115
										l116:
											position, tokenIndex, depth = position116, tokenIndex116, depth116
										}
										if !_rules[rulePyStringQuote]() {
											goto l114
										}
										goto l113
									l114:
										position, tokenIndex, depth = position114, tokenIndex114, depth114
									}
									{
										position117 := position
										depth++
										{
											position118 := position
											depth++
											{
												position119 := position
												depth++
											l120:
												{
													position121, tokenIndex121, depth121 := position, tokenIndex, depth
													{
														position122, tokenIndex122, depth122 := position, tokenIndex, depth
														if buffer[position] != rune('\n') {
															goto l122
														}
														position++
														goto l121
													l122:
														position, tokenIndex, depth = position122, tokenIndex122, depth122
													}
													if !matchDot() {
														goto l121
													}
													goto l120
												l121:
													position, tokenIndex, depth = position121, tokenIndex121, depth121
												}
												depth--
												add(ruleUntilNL, position119)
											}
											depth--
											add(rulePegText, position118)
										}
										if !_rules[ruleNL]() {
											goto l113
										}
										{
											add(ruleAction23, position)
										}
										depth--
										add(rulePyStringLine, position117)
									}
									goto l112
								l113:
									position, tokenIndex, depth = position113, tokenIndex113, depth113
								}
							l124:
								{
									position125, tokenIndex125, depth125 := position, tokenIndex, depth
									if !_rules[ruleWS]() {
										goto l125
									}
									goto l124
								l125:
									position, tokenIndex, depth = position125, tokenIndex125, depth125
								}
								if !_rules[rulePyStringQuote]() {
									goto l98
								}
								if !_rules[ruleLineEnd]() {
									goto l98
								}
								{
									add(ruleAction22, position)
								}
								depth--
								add(rulePyString, position103)
							}
						}
					l101:
						depth--
						add(ruleStepArgument, position100)
					}
					goto l99
				l98:
					position, tokenIndex, depth = position98, tokenIndex98, depth98
				}
			l99:
				{
					add(ruleAction20, position)
				}
				depth--
				add(ruleStep, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 7 StepArgument <- <(Table / PyString)> */
		nil,
		/* 8 PyString <- <((WS* NL)* <WS*> PyStringQuote NL Action21 (!(WS* PyStringQuote) PyStringLine)* WS* PyStringQuote LineEnd Action22)> */
		nil,
		/* 9 PyStringQuote <- <('"' '"' '"')> */
		func() bool {
			position130, tokenIndex130, depth130 := position, tokenIndex, depth
			{
				position131 := position
				depth++
				if buffer[position] != rune('"') {
					goto l130
				}
				position++
				if buffer[position] != rune('"') {
					goto l130
				}
				position++
				if buffer[position] != rune('"') {
					goto l130
				}
				position++
				depth--
				add(rulePyStringQuote, position131)
			}
			return true
		l130:
			position, tokenIndex, depth = position130, tokenIndex130, depth130
			return false
		},
		/* 10 PyStringLine <- <(<UntilNL> NL Action23)> */
		nil,
		/* 11 Table <- <(Action24 TableRow+ Action25)> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				{
					add(ruleAction24, position)
				}
				{
					position138 := position
					depth++
					{
						add(ruleAction26, position)
					}
					if !_rules[ruleOS]() {
						goto l133
					}
					if buffer[position] != rune('|') {
						goto l133
					}
					position++
					{
						position142 := position
						depth++
						{
							position143 := position
							depth++
							{
								position146, tokenIndex146, depth146 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '|':
										if buffer[position] != rune('|') {
											goto l146
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l146
										}
										position++
										break
									default:
										if buffer[position] != rune('\r') {
											goto l146
										}
										position++
										break
									}
								}

								goto l133
							l146:
								position, tokenIndex, depth = position146, tokenIndex146, depth146
							}
							if !matchDot() {
								goto l133
							}
						l144:
							{
								position145, tokenIndex145, depth145 := position, tokenIndex, depth
								{
									position148, tokenIndex148, depth148 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l148
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l148
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l148
											}
											position++
											break
										}
									}

									goto l145
								l148:
									position, tokenIndex, depth = position148, tokenIndex148, depth148
								}
								if !matchDot() {
									goto l145
								}
								goto l144
							l145:
								position, tokenIndex, depth = position145, tokenIndex145, depth145
							}
							depth--
							add(rulePegText, position143)
						}
						if buffer[position] != rune('|') {
							goto l133
						}
						position++
						{
							add(ruleAction28, position)
						}
						depth--
						add(ruleTableCell, position142)
					}
				l140:
					{
						position141, tokenIndex141, depth141 := position, tokenIndex, depth
						{
							position151 := position
							depth++
							{
								position152 := position
								depth++
								{
									position155, tokenIndex155, depth155 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l155
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l155
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l155
											}
											position++
											break
										}
									}

									goto l141
								l155:
									position, tokenIndex, depth = position155, tokenIndex155, depth155
								}
								if !matchDot() {
									goto l141
								}
							l153:
								{
									position154, tokenIndex154, depth154 := position, tokenIndex, depth
									{
										position157, tokenIndex157, depth157 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l157
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l157
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l157
												}
												position++
												break
											}
										}

										goto l154
									l157:
										position, tokenIndex, depth = position157, tokenIndex157, depth157
									}
									if !matchDot() {
										goto l154
									}
									goto l153
								l154:
									position, tokenIndex, depth = position154, tokenIndex154, depth154
								}
								depth--
								add(rulePegText, position152)
							}
							if buffer[position] != rune('|') {
								goto l141
							}
							position++
							{
								add(ruleAction28, position)
							}
							depth--
							add(ruleTableCell, position151)
						}
						goto l140
					l141:
						position, tokenIndex, depth = position141, tokenIndex141, depth141
					}
					if !_rules[ruleLineEnd]() {
						goto l133
					}
					{
						add(ruleAction27, position)
					}
					depth--
					add(ruleTableRow, position138)
				}
			l136:
				{
					position137, tokenIndex137, depth137 := position, tokenIndex, depth
					{
						position161 := position
						depth++
						{
							add(ruleAction26, position)
						}
						if !_rules[ruleOS]() {
							goto l137
						}
						if buffer[position] != rune('|') {
							goto l137
						}
						position++
						{
							position165 := position
							depth++
							{
								position166 := position
								depth++
								{
									position169, tokenIndex169, depth169 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l169
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l169
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l169
											}
											position++
											break
										}
									}

									goto l137
								l169:
									position, tokenIndex, depth = position169, tokenIndex169, depth169
								}
								if !matchDot() {
									goto l137
								}
							l167:
								{
									position168, tokenIndex168, depth168 := position, tokenIndex, depth
									{
										position171, tokenIndex171, depth171 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l171
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l171
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l171
												}
												position++
												break
											}
										}

										goto l168
									l171:
										position, tokenIndex, depth = position171, tokenIndex171, depth171
									}
									if !matchDot() {
										goto l168
									}
									goto l167
								l168:
									position, tokenIndex, depth = position168, tokenIndex168, depth168
								}
								depth--
								add(rulePegText, position166)
							}
							if buffer[position] != rune('|') {
								goto l137
							}
							position++
							{
								add(ruleAction28, position)
							}
							depth--
							add(ruleTableCell, position165)
						}
					l163:
						{
							position164, tokenIndex164, depth164 := position, tokenIndex, depth
							{
								position174 := position
								depth++
								{
									position175 := position
									depth++
									{
										position178, tokenIndex178, depth178 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l178
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l178
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l178
												}
												position++
												break
											}
										}

										goto l164
									l178:
										position, tokenIndex, depth = position178, tokenIndex178, depth178
									}
									if !matchDot() {
										goto l164
									}
								l176:
									{
										position177, tokenIndex177, depth177 := position, tokenIndex, depth
										{
											position180, tokenIndex180, depth180 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '|':
													if buffer[position] != rune('|') {
														goto l180
													}
													position++
													break
												case '\n':
													if buffer[position] != rune('\n') {
														goto l180
													}
													position++
													break
												default:
													if buffer[position] != rune('\r') {
														goto l180
													}
													position++
													break
												}
											}

											goto l177
										l180:
											position, tokenIndex, depth = position180, tokenIndex180, depth180
										}
										if !matchDot() {
											goto l177
										}
										goto l176
									l177:
										position, tokenIndex, depth = position177, tokenIndex177, depth177
									}
									depth--
									add(rulePegText, position175)
								}
								if buffer[position] != rune('|') {
									goto l164
								}
								position++
								{
									add(ruleAction28, position)
								}
								depth--
								add(ruleTableCell, position174)
							}
							goto l163
						l164:
							position, tokenIndex, depth = position164, tokenIndex164, depth164
						}
						if !_rules[ruleLineEnd]() {
							goto l137
						}
						{
							add(ruleAction27, position)
						}
						depth--
						add(ruleTableRow, position161)
					}
					goto l136
				l137:
					position, tokenIndex, depth = position137, tokenIndex137, depth137
				}
				{
					add(ruleAction25, position)
				}
				depth--
				add(ruleTable, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 12 TableRow <- <(Action26 OS '|' TableCell+ LineEnd Action27)> */
		nil,
		/* 13 TableCell <- <(<(!((&('|') '|') | (&('\n') '\n') | (&('\r') '\r')) .)+> '|' Action28)> */
		nil,
		/* 14 Tags <- <((Tag+ WS* LineEnd?)* OS)> */
		func() bool {
			position187, tokenIndex187, depth187 := position, tokenIndex, depth
			{
				position188 := position
				depth++
			l189:
				{
					position190, tokenIndex190, depth190 := position, tokenIndex, depth
					{
						position193 := position
						depth++
						if !_rules[ruleOS]() {
							goto l190
						}
						if buffer[position] != rune('@') {
							goto l190
						}
						position++
						{
							position194 := position
							depth++
							if !_rules[ruleWord]() {
								goto l190
							}
							depth--
							add(rulePegText, position194)
						}
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleTag, position193)
					}
				l191:
					{
						position192, tokenIndex192, depth192 := position, tokenIndex, depth
						{
							position196 := position
							depth++
							if !_rules[ruleOS]() {
								goto l192
							}
							if buffer[position] != rune('@') {
								goto l192
							}
							position++
							{
								position197 := position
								depth++
								if !_rules[ruleWord]() {
									goto l192
								}
								depth--
								add(rulePegText, position197)
							}
							{
								add(ruleAction29, position)
							}
							depth--
							add(ruleTag, position196)
						}
						goto l191
					l192:
						position, tokenIndex, depth = position192, tokenIndex192, depth192
					}
				l199:
					{
						position200, tokenIndex200, depth200 := position, tokenIndex, depth
						if !_rules[ruleWS]() {
							goto l200
						}
						goto l199
					l200:
						position, tokenIndex, depth = position200, tokenIndex200, depth200
					}
					{
						position201, tokenIndex201, depth201 := position, tokenIndex, depth
						if !_rules[ruleLineEnd]() {
							goto l201
						}
						goto l202
					l201:
						position, tokenIndex, depth = position201, tokenIndex201, depth201
					}
				l202:
					goto l189
				l190:
					position, tokenIndex, depth = position190, tokenIndex190, depth190
				}
				if !_rules[ruleOS]() {
					goto l187
				}
				depth--
				add(ruleTags, position188)
			}
			return true
		l187:
			position, tokenIndex, depth = position187, tokenIndex187, depth187
			return false
		},
		/* 15 Tag <- <(OS '@' <Word> Action29)> */
		nil,
		/* 16 Word <- <(!((&('#') '#') | (&('"') '"') | (&(' ') ' ') | (&('\t') '\t') | (&('\n') '\n') | (&('\r') '\r')) .)+> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				{
					position208, tokenIndex208, depth208 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '#':
							if buffer[position] != rune('#') {
								goto l208
							}
							position++
							break
						case '"':
							if buffer[position] != rune('"') {
								goto l208
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l208
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l208
							}
							position++
							break
						case '\n':
							if buffer[position] != rune('\n') {
								goto l208
							}
							position++
							break
						default:
							if buffer[position] != rune('\r') {
								goto l208
							}
							position++
							break
						}
					}

					goto l204
				l208:
					position, tokenIndex, depth = position208, tokenIndex208, depth208
				}
				if !matchDot() {
					goto l204
				}
			l206:
				{
					position207, tokenIndex207, depth207 := position, tokenIndex, depth
					{
						position210, tokenIndex210, depth210 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '#':
								if buffer[position] != rune('#') {
									goto l210
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l210
								}
								position++
								break
							case ' ':
								if buffer[position] != rune(' ') {
									goto l210
								}
								position++
								break
							case '\t':
								if buffer[position] != rune('\t') {
									goto l210
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l210
								}
								position++
								break
							default:
								if buffer[position] != rune('\r') {
									goto l210
								}
								position++
								break
							}
						}

						goto l207
					l210:
						position, tokenIndex, depth = position210, tokenIndex210, depth210
					}
					if !matchDot() {
						goto l207
					}
					goto l206
				l207:
					position, tokenIndex, depth = position207, tokenIndex207, depth207
				}
				depth--
				add(ruleWord, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 17 EscapedChar <- <('\\' .)> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l212
				}
				position++
				if !matchDot() {
					goto l212
				}
				depth--
				add(ruleEscapedChar, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 18 QuotedString <- <('"' (EscapedChar / (!((&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+)* '"')> */
		nil,
		/* 19 UntilLineEnd <- <(EscapedChar / (!((&('#') '#') | (&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+ / QuotedString)+> */
		func() bool {
			position215, tokenIndex215, depth215 := position, tokenIndex, depth
			{
				position216 := position
				depth++
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if !_rules[ruleEscapedChar]() {
						goto l220
					}
					goto l219
				l220:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
					{
						position224, tokenIndex224, depth224 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '#':
								if buffer[position] != rune('#') {
									goto l224
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l224
								}
								position++
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l224
								}
								position++
								break
							default:
								if buffer[position] != rune('\n') {
									goto l224
								}
								position++
								break
							}
						}

						goto l221
					l224:
						position, tokenIndex, depth = position224, tokenIndex224, depth224
					}
					if !matchDot() {
						goto l221
					}
				l222:
					{
						position223, tokenIndex223, depth223 := position, tokenIndex, depth
						{
							position226, tokenIndex226, depth226 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l226
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l226
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l226
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l226
									}
									position++
									break
								}
							}

							goto l223
						l226:
							position, tokenIndex, depth = position226, tokenIndex226, depth226
						}
						if !matchDot() {
							goto l223
						}
						goto l222
					l223:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
					}
					goto l219
				l221:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
					{
						position228 := position
						depth++
						if buffer[position] != rune('"') {
							goto l215
						}
						position++
					l229:
						{
							position230, tokenIndex230, depth230 := position, tokenIndex, depth
							{
								position231, tokenIndex231, depth231 := position, tokenIndex, depth
								if !_rules[ruleEscapedChar]() {
									goto l232
								}
								goto l231
							l232:
								position, tokenIndex, depth = position231, tokenIndex231, depth231
								{
									position235, tokenIndex235, depth235 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '"':
											if buffer[position] != rune('"') {
												goto l235
											}
											position++
											break
										case '\\':
											if buffer[position] != rune('\\') {
												goto l235
											}
											position++
											break
										default:
											if buffer[position] != rune('\n') {
												goto l235
											}
											position++
											break
										}
									}

									goto l230
								l235:
									position, tokenIndex, depth = position235, tokenIndex235, depth235
								}
								if !matchDot() {
									goto l230
								}
							l233:
								{
									position234, tokenIndex234, depth234 := position, tokenIndex, depth
									{
										position237, tokenIndex237, depth237 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l237
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l237
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l237
												}
												position++
												break
											}
										}

										goto l234
									l237:
										position, tokenIndex, depth = position237, tokenIndex237, depth237
									}
									if !matchDot() {
										goto l234
									}
									goto l233
								l234:
									position, tokenIndex, depth = position234, tokenIndex234, depth234
								}
							}
						l231:
							goto l229
						l230:
							position, tokenIndex, depth = position230, tokenIndex230, depth230
						}
						if buffer[position] != rune('"') {
							goto l215
						}
						position++
						depth--
						add(ruleQuotedString, position228)
					}
				}
			l219:
			l217:
				{
					position218, tokenIndex218, depth218 := position, tokenIndex, depth
					{
						position239, tokenIndex239, depth239 := position, tokenIndex, depth
						if !_rules[ruleEscapedChar]() {
							goto l240
						}
						goto l239
					l240:
						position, tokenIndex, depth = position239, tokenIndex239, depth239
						{
							position244, tokenIndex244, depth244 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l244
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l244
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l244
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l244
									}
									position++
									break
								}
							}

							goto l241
						l244:
							position, tokenIndex, depth = position244, tokenIndex244, depth244
						}
						if !matchDot() {
							goto l241
						}
					l242:
						{
							position243, tokenIndex243, depth243 := position, tokenIndex, depth
							{
								position246, tokenIndex246, depth246 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '#':
										if buffer[position] != rune('#') {
											goto l246
										}
										position++
										break
									case '"':
										if buffer[position] != rune('"') {
											goto l246
										}
										position++
										break
									case '\\':
										if buffer[position] != rune('\\') {
											goto l246
										}
										position++
										break
									default:
										if buffer[position] != rune('\n') {
											goto l246
										}
										position++
										break
									}
								}

								goto l243
							l246:
								position, tokenIndex, depth = position246, tokenIndex246, depth246
							}
							if !matchDot() {
								goto l243
							}
							goto l242
						l243:
							position, tokenIndex, depth = position243, tokenIndex243, depth243
						}
						goto l239
					l241:
						position, tokenIndex, depth = position239, tokenIndex239, depth239
						{
							position248 := position
							depth++
							if buffer[position] != rune('"') {
								goto l218
							}
							position++
						l249:
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								{
									position251, tokenIndex251, depth251 := position, tokenIndex, depth
									if !_rules[ruleEscapedChar]() {
										goto l252
									}
									goto l251
								l252:
									position, tokenIndex, depth = position251, tokenIndex251, depth251
									{
										position255, tokenIndex255, depth255 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l255
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l255
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l255
												}
												position++
												break
											}
										}

										goto l250
									l255:
										position, tokenIndex, depth = position255, tokenIndex255, depth255
									}
									if !matchDot() {
										goto l250
									}
								l253:
									{
										position254, tokenIndex254, depth254 := position, tokenIndex, depth
										{
											position257, tokenIndex257, depth257 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '"':
													if buffer[position] != rune('"') {
														goto l257
													}
													position++
													break
												case '\\':
													if buffer[position] != rune('\\') {
														goto l257
													}
													position++
													break
												default:
													if buffer[position] != rune('\n') {
														goto l257
													}
													position++
													break
												}
											}

											goto l254
										l257:
											position, tokenIndex, depth = position257, tokenIndex257, depth257
										}
										if !matchDot() {
											goto l254
										}
										goto l253
									l254:
										position, tokenIndex, depth = position254, tokenIndex254, depth254
									}
								}
							l251:
								goto l249
							l250:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
							}
							if buffer[position] != rune('"') {
								goto l218
							}
							position++
							depth--
							add(ruleQuotedString, position248)
						}
					}
				l239:
					goto l217
				l218:
					position, tokenIndex, depth = position218, tokenIndex218, depth218
				}
				depth--
				add(ruleUntilLineEnd, position216)
			}
			return true
		l215:
			position, tokenIndex, depth = position215, tokenIndex215, depth215
			return false
		},
		/* 20 LineEnd <- <(WS* LineComment? NL)> */
		func() bool {
			position259, tokenIndex259, depth259 := position, tokenIndex, depth
			{
				position260 := position
				depth++
			l261:
				{
					position262, tokenIndex262, depth262 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l262
					}
					goto l261
				l262:
					position, tokenIndex, depth = position262, tokenIndex262, depth262
				}
				{
					position263, tokenIndex263, depth263 := position, tokenIndex, depth
					if !_rules[ruleLineComment]() {
						goto l263
					}
					goto l264
				l263:
					position, tokenIndex, depth = position263, tokenIndex263, depth263
				}
			l264:
				if !_rules[ruleNL]() {
					goto l259
				}
				depth--
				add(ruleLineEnd, position260)
			}
			return true
		l259:
			position, tokenIndex, depth = position259, tokenIndex259, depth259
			return false
		},
		/* 21 LineComment <- <('#' <(!'\n' .)*> Action30)> */
		func() bool {
			position265, tokenIndex265, depth265 := position, tokenIndex, depth
			{
				position266 := position
				depth++
				if buffer[position] != rune('#') {
					goto l265
				}
				position++
				{
					position267 := position
					depth++
				l268:
					{
						position269, tokenIndex269, depth269 := position, tokenIndex, depth
						{
							position270, tokenIndex270, depth270 := position, tokenIndex, depth
							if buffer[position] != rune('\n') {
								goto l270
							}
							position++
							goto l269
						l270:
							position, tokenIndex, depth = position270, tokenIndex270, depth270
						}
						if !matchDot() {
							goto l269
						}
						goto l268
					l269:
						position, tokenIndex, depth = position269, tokenIndex269, depth269
					}
					depth--
					add(rulePegText, position267)
				}
				{
					add(ruleAction30, position)
				}
				depth--
				add(ruleLineComment, position266)
			}
			return true
		l265:
			position, tokenIndex, depth = position265, tokenIndex265, depth265
			return false
		},
		/* 22 BlankLine <- <(((WS LineEnd) / (LineComment? NL)) Action31)> */
		func() bool {
			position272, tokenIndex272, depth272 := position, tokenIndex, depth
			{
				position273 := position
				depth++
				{
					position274, tokenIndex274, depth274 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l275
					}
					if !_rules[ruleLineEnd]() {
						goto l275
					}
					goto l274
				l275:
					position, tokenIndex, depth = position274, tokenIndex274, depth274
					{
						position276, tokenIndex276, depth276 := position, tokenIndex, depth
						if !_rules[ruleLineComment]() {
							goto l276
						}
						goto l277
					l276:
						position, tokenIndex, depth = position276, tokenIndex276, depth276
					}
				l277:
					if !_rules[ruleNL]() {
						goto l272
					}
				}
			l274:
				{
					add(ruleAction31, position)
				}
				depth--
				add(ruleBlankLine, position273)
			}
			return true
		l272:
			position, tokenIndex, depth = position272, tokenIndex272, depth272
			return false
		},
		/* 23 OS <- <(NL / WS)*> */
		func() bool {
			{
				position280 := position
				depth++
			l281:
				{
					position282, tokenIndex282, depth282 := position, tokenIndex, depth
					{
						position283, tokenIndex283, depth283 := position, tokenIndex, depth
						if !_rules[ruleNL]() {
							goto l284
						}
						goto l283
					l284:
						position, tokenIndex, depth = position283, tokenIndex283, depth283
						if !_rules[ruleWS]() {
							goto l282
						}
					}
				l283:
					goto l281
				l282:
					position, tokenIndex, depth = position282, tokenIndex282, depth282
				}
				depth--
				add(ruleOS, position280)
			}
			return true
		},
		/* 24 WS <- <(' ' / '\t')> */
		func() bool {
			position285, tokenIndex285, depth285 := position, tokenIndex, depth
			{
				position286 := position
				depth++
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l288
					}
					position++
					goto l287
				l288:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
					if buffer[position] != rune('\t') {
						goto l285
					}
					position++
				}
			l287:
				depth--
				add(ruleWS, position286)
			}
			return true
		l285:
			position, tokenIndex, depth = position285, tokenIndex285, depth285
			return false
		},
		/* 25 UntilNL <- <(!'\n' .)*> */
		nil,
		/* 26 NL <- <('\n' / '\r' / ('\r' '\n'))> */
		func() bool {
			position290, tokenIndex290, depth290 := position, tokenIndex, depth
			{
				position291 := position
				depth++
				{
					position292, tokenIndex292, depth292 := position, tokenIndex, depth
					if buffer[position] != rune('\n') {
						goto l293
					}
					position++
					goto l292
				l293:
					position, tokenIndex, depth = position292, tokenIndex292, depth292
					if buffer[position] != rune('\r') {
						goto l294
					}
					position++
					goto l292
				l294:
					position, tokenIndex, depth = position292, tokenIndex292, depth292
					if buffer[position] != rune('\r') {
						goto l290
					}
					position++
					if buffer[position] != rune('\n') {
						goto l290
					}
					position++
				}
			l292:
				depth--
				add(ruleNL, position291)
			}
			return true
		l290:
			position, tokenIndex, depth = position290, tokenIndex290, depth290
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
		/* 44 Action15 <- <{ p.beginOutlineExamples() }> */
		nil,
		/* 45 Action16 <- <{ p.endOutlineExamples() }> */
		nil,
		/* 46 Action17 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 47 Action18 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 48 Action19 <- <{ p.beginStep(trimWS(p.buf1), trimWS(p.buf2)) }> */
		nil,
		/* 49 Action20 <- <{ p.endStep() }> */
		nil,
		/* 50 Action21 <- <{ p.beginPyString(buffer[begin:end]) }> */
		nil,
		/* 51 Action22 <- <{ p.endPyString() }> */
		nil,
		/* 52 Action23 <- <{ p.bufferPyString(buffer[begin:end]) }> */
		nil,
		/* 53 Action24 <- <{ p.beginTable() }> */
		nil,
		/* 54 Action25 <- <{ p.endTable() }> */
		nil,
		/* 55 Action26 <- <{ p.beginTableRow() }> */
		nil,
		/* 56 Action27 <- <{ p.endTableRow() }> */
		nil,
		/* 57 Action28 <- <{ p.beginTableCell(); p.endTableCell(trimWS(buffer[begin:end])) }> */
		nil,
		/* 58 Action29 <- <{ p.buftags = append(p.buftags, buffer[begin:end]) }> */
		nil,
		/* 59 Action30 <- <{ p.bufcmt = buffer[begin:end]; p.triggerComment(p.bufcmt) }> */
		nil,
		/* 60 Action31 <- <{ p.triggerBlankLine() }> */
		nil,
	}
	p.rules = _rules
}
