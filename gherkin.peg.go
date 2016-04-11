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
	ruleScenarioKeyWord
	ruleStepKeyWord
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
	ruleAction33
	ruleAction34
	ruleAction35
	ruleAction36
	ruleAction37
	ruleAction38
	ruleAction39
	ruleAction40
	ruleAction41

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"Begin",
	"ScenarioKeyWord",
	"StepKeyWord",
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
	"Action33",
	"Action34",
	"Action35",
	"Action36",
	"Action37",
	"Action38",
	"Action39",
	"Action40",
	"Action41",

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
	rules  [73]func() bool
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
			p.buf2 = buffer[begin:end]
		case ruleAction8:
			p.buf2 = p.buf2 + buffer[begin:end]
		case ruleAction9:
			p.buf2 = p.buf2 + "\n"
		case ruleAction10:
			p.beginBackground(trimWS(p.buf1), trimWSML(p.buf2), p.buftags)
			p.buftags = nil
		case ruleAction11:
			p.endBackground()
		case ruleAction12:
			p.buf1 = buffer[begin:end]
		case ruleAction13:
			p.buf2 = buffer[begin:end]
		case ruleAction14:
			p.buf2 = p.buf2 + buffer[begin:end]
		case ruleAction15:
			p.buf2 = p.buf2 + "\n"
		case ruleAction16:
			p.beginScenario(trimWS(p.buf1), trimWSML(p.buf2), p.buftags)
			p.buftags = nil
		case ruleAction17:
			p.endScenario()
		case ruleAction18:
			p.buf1 = buffer[begin:end]
		case ruleAction19:
			p.buf2 = buffer[begin:end]
		case ruleAction20:
			p.buf2 = p.buf2 + buffer[begin:end]
		case ruleAction21:
			p.buf2 = p.buf2 + "\n"
		case ruleAction22:
			p.beginOutline(trimWS(p.buf1), trimWSML(p.buf2), p.buftags)
			p.buftags = nil
		case ruleAction23:
			p.endOutline()
		case ruleAction24:
			p.buf1 = buffer[begin:end]
		case ruleAction25:
			p.beginOutlineExamples(trimWS(p.buf1))
		case ruleAction26:
			p.endOutlineExamples()
		case ruleAction27:
			p.buf1 = buffer[begin:end]
		case ruleAction28:
			p.buf2 = buffer[begin:end]
		case ruleAction29:
			p.beginStep(trimWS(p.buf1), trimWS(p.buf2))
		case ruleAction30:
			p.endStep()
		case ruleAction31:
			p.beginPyString(buffer[begin:end])
		case ruleAction32:
			p.endPyString()
		case ruleAction33:
			p.bufferPyString(buffer[begin:end])
		case ruleAction34:
			p.beginTable()
		case ruleAction35:
			p.endTable()
		case ruleAction36:
			p.beginTableRow()
		case ruleAction37:
			p.endTableRow()
		case ruleAction38:
			p.beginTableCell()
			p.endTableCell(trimWS(buffer[begin:end]))
		case ruleAction39:
			p.buftags = append(p.buftags, buffer[begin:end])
		case ruleAction40:
			p.bufcmt = buffer[begin:end]
			p.triggerComment(p.bufcmt)
		case ruleAction41:
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
									if buffer[position] != rune('@') {
										goto l19
									}
									position++
									if !_rules[ruleWord]() {
										goto l19
									}
									goto l18
								l19:
									position, tokenIndex, depth = position18, tokenIndex18, depth18
									if !_rules[ruleScenarioKeyWord]() {
										goto l17
									}
								}
							l18:
								goto l14
							l17:
								position, tokenIndex, depth = position17, tokenIndex17, depth17
							}
							{
								position20 := position
								depth++
								{
									position21, tokenIndex21, depth21 := position, tokenIndex, depth
									if !_rules[ruleUntilLineEnd]() {
										goto l21
									}
									goto l22
								l21:
									position, tokenIndex, depth = position21, tokenIndex21, depth21
								}
							l22:
								depth--
								add(rulePegText, position20)
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
					l26:
						{
							position27, tokenIndex27, depth27 := position, tokenIndex, depth
							{
								position28, tokenIndex28, depth28 := position, tokenIndex, depth
								{
									position30 := position
									depth++
									if !_rules[ruleTags]() {
										goto l29
									}
									if buffer[position] != rune('B') {
										goto l29
									}
									position++
									if buffer[position] != rune('a') {
										goto l29
									}
									position++
									if buffer[position] != rune('c') {
										goto l29
									}
									position++
									if buffer[position] != rune('k') {
										goto l29
									}
									position++
									if buffer[position] != rune('g') {
										goto l29
									}
									position++
									if buffer[position] != rune('r') {
										goto l29
									}
									position++
									if buffer[position] != rune('o') {
										goto l29
									}
									position++
									if buffer[position] != rune('u') {
										goto l29
									}
									position++
									if buffer[position] != rune('n') {
										goto l29
									}
									position++
									if buffer[position] != rune('d') {
										goto l29
									}
									position++
									if buffer[position] != rune(':') {
										goto l29
									}
									position++
								l31:
									{
										position32, tokenIndex32, depth32 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l32
										}
										goto l31
									l32:
										position, tokenIndex, depth = position32, tokenIndex32, depth32
									}
									{
										position33 := position
										depth++
										{
											position34, tokenIndex34, depth34 := position, tokenIndex, depth
											if !_rules[ruleUntilLineEnd]() {
												goto l34
											}
											goto l35
										l34:
											position, tokenIndex, depth = position34, tokenIndex34, depth34
										}
									l35:
										depth--
										add(rulePegText, position33)
									}
									{
										add(ruleAction6, position)
									}
									{
										position37 := position
										depth++
										depth--
										add(rulePegText, position37)
									}
									{
										add(ruleAction7, position)
									}
									if !_rules[ruleLineEnd]() {
										goto l29
									}
								l39:
									{
										position40, tokenIndex40, depth40 := position, tokenIndex, depth
									l41:
										{
											position42, tokenIndex42, depth42 := position, tokenIndex, depth
											if !_rules[ruleWS]() {
												goto l42
											}
											goto l41
										l42:
											position, tokenIndex, depth = position42, tokenIndex42, depth42
										}
										{
											position43, tokenIndex43, depth43 := position, tokenIndex, depth
											{
												position44, tokenIndex44, depth44 := position, tokenIndex, depth
												if buffer[position] != rune('@') {
													goto l45
												}
												position++
												if !_rules[ruleWord]() {
													goto l45
												}
												goto l44
											l45:
												position, tokenIndex, depth = position44, tokenIndex44, depth44
												if !_rules[ruleScenarioKeyWord]() {
													goto l46
												}
												goto l44
											l46:
												position, tokenIndex, depth = position44, tokenIndex44, depth44
												if !_rules[ruleStepKeyWord]() {
													goto l43
												}
											}
										l44:
											goto l40
										l43:
											position, tokenIndex, depth = position43, tokenIndex43, depth43
										}
										{
											position47 := position
											depth++
											{
												position48, tokenIndex48, depth48 := position, tokenIndex, depth
												if !_rules[ruleUntilLineEnd]() {
													goto l48
												}
												goto l49
											l48:
												position, tokenIndex, depth = position48, tokenIndex48, depth48
											}
										l49:
											depth--
											add(rulePegText, position47)
										}
										{
											add(ruleAction8, position)
										}
										if !_rules[ruleLineEnd]() {
											goto l40
										}
										{
											add(ruleAction9, position)
										}
										goto l39
									l40:
										position, tokenIndex, depth = position40, tokenIndex40, depth40
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
									add(ruleBackground, position30)
								}
								goto l28
							l29:
								position, tokenIndex, depth = position28, tokenIndex28, depth28
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
									{
										position66 := position
										depth++
										depth--
										add(rulePegText, position66)
									}
									{
										add(ruleAction13, position)
									}
									if !_rules[ruleLineEnd]() {
										goto l58
									}
								l68:
									{
										position69, tokenIndex69, depth69 := position, tokenIndex, depth
									l70:
										{
											position71, tokenIndex71, depth71 := position, tokenIndex, depth
											if !_rules[ruleWS]() {
												goto l71
											}
											goto l70
										l71:
											position, tokenIndex, depth = position71, tokenIndex71, depth71
										}
										{
											position72, tokenIndex72, depth72 := position, tokenIndex, depth
											{
												position73, tokenIndex73, depth73 := position, tokenIndex, depth
												if buffer[position] != rune('@') {
													goto l74
												}
												position++
												if !_rules[ruleWord]() {
													goto l74
												}
												goto l73
											l74:
												position, tokenIndex, depth = position73, tokenIndex73, depth73
												if !_rules[ruleScenarioKeyWord]() {
													goto l75
												}
												goto l73
											l75:
												position, tokenIndex, depth = position73, tokenIndex73, depth73
												if !_rules[ruleStepKeyWord]() {
													goto l72
												}
											}
										l73:
											goto l69
										l72:
											position, tokenIndex, depth = position72, tokenIndex72, depth72
										}
										{
											position76 := position
											depth++
											{
												position77, tokenIndex77, depth77 := position, tokenIndex, depth
												if !_rules[ruleUntilLineEnd]() {
													goto l77
												}
												goto l78
											l77:
												position, tokenIndex, depth = position77, tokenIndex77, depth77
											}
										l78:
											depth--
											add(rulePegText, position76)
										}
										{
											add(ruleAction14, position)
										}
										if !_rules[ruleLineEnd]() {
											goto l69
										}
										{
											add(ruleAction15, position)
										}
										goto l68
									l69:
										position, tokenIndex, depth = position69, tokenIndex69, depth69
									}
									{
										add(ruleAction16, position)
									}
								l82:
									{
										position83, tokenIndex83, depth83 := position, tokenIndex, depth
										{
											position84, tokenIndex84, depth84 := position, tokenIndex, depth
											if !_rules[ruleStep]() {
												goto l85
											}
											goto l84
										l85:
											position, tokenIndex, depth = position84, tokenIndex84, depth84
											if !_rules[ruleBlankLine]() {
												goto l83
											}
										}
									l84:
										goto l82
									l83:
										position, tokenIndex, depth = position83, tokenIndex83, depth83
									}
									{
										add(ruleAction17, position)
									}
									depth--
									add(ruleScenario, position59)
								}
								goto l28
							l58:
								position, tokenIndex, depth = position28, tokenIndex28, depth28
								{
									position88 := position
									depth++
									if !_rules[ruleTags]() {
										goto l87
									}
									if buffer[position] != rune('S') {
										goto l87
									}
									position++
									if buffer[position] != rune('c') {
										goto l87
									}
									position++
									if buffer[position] != rune('e') {
										goto l87
									}
									position++
									if buffer[position] != rune('n') {
										goto l87
									}
									position++
									if buffer[position] != rune('a') {
										goto l87
									}
									position++
									if buffer[position] != rune('r') {
										goto l87
									}
									position++
									if buffer[position] != rune('i') {
										goto l87
									}
									position++
									if buffer[position] != rune('o') {
										goto l87
									}
									position++
									if buffer[position] != rune(' ') {
										goto l87
									}
									position++
									if buffer[position] != rune('O') {
										goto l87
									}
									position++
									if buffer[position] != rune('u') {
										goto l87
									}
									position++
									if buffer[position] != rune('t') {
										goto l87
									}
									position++
									if buffer[position] != rune('l') {
										goto l87
									}
									position++
									if buffer[position] != rune('i') {
										goto l87
									}
									position++
									if buffer[position] != rune('n') {
										goto l87
									}
									position++
									if buffer[position] != rune('e') {
										goto l87
									}
									position++
									if buffer[position] != rune(':') {
										goto l87
									}
									position++
								l89:
									{
										position90, tokenIndex90, depth90 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l90
										}
										goto l89
									l90:
										position, tokenIndex, depth = position90, tokenIndex90, depth90
									}
									{
										position91 := position
										depth++
										{
											position92, tokenIndex92, depth92 := position, tokenIndex, depth
											if !_rules[ruleUntilLineEnd]() {
												goto l92
											}
											goto l93
										l92:
											position, tokenIndex, depth = position92, tokenIndex92, depth92
										}
									l93:
										depth--
										add(rulePegText, position91)
									}
									{
										add(ruleAction18, position)
									}
									{
										position95 := position
										depth++
										depth--
										add(rulePegText, position95)
									}
									{
										add(ruleAction19, position)
									}
									if !_rules[ruleLineEnd]() {
										goto l87
									}
								l97:
									{
										position98, tokenIndex98, depth98 := position, tokenIndex, depth
									l99:
										{
											position100, tokenIndex100, depth100 := position, tokenIndex, depth
											if !_rules[ruleWS]() {
												goto l100
											}
											goto l99
										l100:
											position, tokenIndex, depth = position100, tokenIndex100, depth100
										}
										{
											position101, tokenIndex101, depth101 := position, tokenIndex, depth
											{
												position102, tokenIndex102, depth102 := position, tokenIndex, depth
												if buffer[position] != rune('@') {
													goto l103
												}
												position++
												if !_rules[ruleWord]() {
													goto l103
												}
												goto l102
											l103:
												position, tokenIndex, depth = position102, tokenIndex102, depth102
												if !_rules[ruleScenarioKeyWord]() {
													goto l104
												}
												goto l102
											l104:
												position, tokenIndex, depth = position102, tokenIndex102, depth102
												if !_rules[ruleStepKeyWord]() {
													goto l101
												}
											}
										l102:
											goto l98
										l101:
											position, tokenIndex, depth = position101, tokenIndex101, depth101
										}
										{
											position105 := position
											depth++
											{
												position106, tokenIndex106, depth106 := position, tokenIndex, depth
												if !_rules[ruleUntilLineEnd]() {
													goto l106
												}
												goto l107
											l106:
												position, tokenIndex, depth = position106, tokenIndex106, depth106
											}
										l107:
											depth--
											add(rulePegText, position105)
										}
										{
											add(ruleAction20, position)
										}
										if !_rules[ruleLineEnd]() {
											goto l98
										}
										{
											add(ruleAction21, position)
										}
										goto l97
									l98:
										position, tokenIndex, depth = position98, tokenIndex98, depth98
									}
									{
										add(ruleAction22, position)
									}
								l111:
									{
										position112, tokenIndex112, depth112 := position, tokenIndex, depth
										{
											position113, tokenIndex113, depth113 := position, tokenIndex, depth
											if !_rules[ruleStep]() {
												goto l114
											}
											goto l113
										l114:
											position, tokenIndex, depth = position113, tokenIndex113, depth113
											if !_rules[ruleBlankLine]() {
												goto l112
											}
										}
									l113:
										goto l111
									l112:
										position, tokenIndex, depth = position112, tokenIndex112, depth112
									}
								l115:
									{
										position116, tokenIndex116, depth116 := position, tokenIndex, depth
										{
											position117, tokenIndex117, depth117 := position, tokenIndex, depth
											{
												position119 := position
												depth++
												if !_rules[ruleOS]() {
													goto l118
												}
												if buffer[position] != rune('E') {
													goto l118
												}
												position++
												if buffer[position] != rune('x') {
													goto l118
												}
												position++
												if buffer[position] != rune('a') {
													goto l118
												}
												position++
												if buffer[position] != rune('m') {
													goto l118
												}
												position++
												if buffer[position] != rune('p') {
													goto l118
												}
												position++
												if buffer[position] != rune('l') {
													goto l118
												}
												position++
												if buffer[position] != rune('e') {
													goto l118
												}
												position++
												if buffer[position] != rune('s') {
													goto l118
												}
												position++
												if buffer[position] != rune(':') {
													goto l118
												}
												position++
											l120:
												{
													position121, tokenIndex121, depth121 := position, tokenIndex, depth
													if !_rules[ruleWS]() {
														goto l121
													}
													goto l120
												l121:
													position, tokenIndex, depth = position121, tokenIndex121, depth121
												}
												{
													position122 := position
													depth++
													{
														position123, tokenIndex123, depth123 := position, tokenIndex, depth
														if !_rules[ruleUntilLineEnd]() {
															goto l123
														}
														goto l124
													l123:
														position, tokenIndex, depth = position123, tokenIndex123, depth123
													}
												l124:
													depth--
													add(rulePegText, position122)
												}
												{
													add(ruleAction24, position)
												}
												if !_rules[ruleLineEnd]() {
													goto l118
												}
												{
													add(ruleAction25, position)
												}
												{
													position127, tokenIndex127, depth127 := position, tokenIndex, depth
													if !_rules[ruleTable]() {
														goto l127
													}
													goto l128
												l127:
													position, tokenIndex, depth = position127, tokenIndex127, depth127
												}
											l128:
												{
													add(ruleAction26, position)
												}
												depth--
												add(ruleOutlineExamples, position119)
											}
											goto l117
										l118:
											position, tokenIndex, depth = position117, tokenIndex117, depth117
											if !_rules[ruleBlankLine]() {
												goto l116
											}
										}
									l117:
										goto l115
									l116:
										position, tokenIndex, depth = position116, tokenIndex116, depth116
									}
									{
										add(ruleAction23, position)
									}
									depth--
									add(ruleOutline, position88)
								}
								goto l28
							l87:
								position, tokenIndex, depth = position28, tokenIndex28, depth28
								if !_rules[ruleBlankLine]() {
									goto l27
								}
							}
						l28:
							goto l26
						l27:
							position, tokenIndex, depth = position27, tokenIndex27, depth27
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
					position132, tokenIndex132, depth132 := position, tokenIndex, depth
					if !matchDot() {
						goto l132
					}
					goto l0
				l132:
					position, tokenIndex, depth = position132, tokenIndex132, depth132
				}
				depth--
				add(ruleBegin, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 ScenarioKeyWord <- <(('B' 'a' 'c' 'k' 'g' 'r' 'o' 'u' 'n' 'd' ':') / ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ':') / ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':'))> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if buffer[position] != rune('B') {
						goto l136
					}
					position++
					if buffer[position] != rune('a') {
						goto l136
					}
					position++
					if buffer[position] != rune('c') {
						goto l136
					}
					position++
					if buffer[position] != rune('k') {
						goto l136
					}
					position++
					if buffer[position] != rune('g') {
						goto l136
					}
					position++
					if buffer[position] != rune('r') {
						goto l136
					}
					position++
					if buffer[position] != rune('o') {
						goto l136
					}
					position++
					if buffer[position] != rune('u') {
						goto l136
					}
					position++
					if buffer[position] != rune('n') {
						goto l136
					}
					position++
					if buffer[position] != rune('d') {
						goto l136
					}
					position++
					if buffer[position] != rune(':') {
						goto l136
					}
					position++
					goto l135
				l136:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if buffer[position] != rune('S') {
						goto l137
					}
					position++
					if buffer[position] != rune('c') {
						goto l137
					}
					position++
					if buffer[position] != rune('e') {
						goto l137
					}
					position++
					if buffer[position] != rune('n') {
						goto l137
					}
					position++
					if buffer[position] != rune('a') {
						goto l137
					}
					position++
					if buffer[position] != rune('r') {
						goto l137
					}
					position++
					if buffer[position] != rune('i') {
						goto l137
					}
					position++
					if buffer[position] != rune('o') {
						goto l137
					}
					position++
					if buffer[position] != rune(':') {
						goto l137
					}
					position++
					goto l135
				l137:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if buffer[position] != rune('S') {
						goto l133
					}
					position++
					if buffer[position] != rune('c') {
						goto l133
					}
					position++
					if buffer[position] != rune('e') {
						goto l133
					}
					position++
					if buffer[position] != rune('n') {
						goto l133
					}
					position++
					if buffer[position] != rune('a') {
						goto l133
					}
					position++
					if buffer[position] != rune('r') {
						goto l133
					}
					position++
					if buffer[position] != rune('i') {
						goto l133
					}
					position++
					if buffer[position] != rune('o') {
						goto l133
					}
					position++
					if buffer[position] != rune(' ') {
						goto l133
					}
					position++
					if buffer[position] != rune('O') {
						goto l133
					}
					position++
					if buffer[position] != rune('u') {
						goto l133
					}
					position++
					if buffer[position] != rune('t') {
						goto l133
					}
					position++
					if buffer[position] != rune('l') {
						goto l133
					}
					position++
					if buffer[position] != rune('i') {
						goto l133
					}
					position++
					if buffer[position] != rune('n') {
						goto l133
					}
					position++
					if buffer[position] != rune('e') {
						goto l133
					}
					position++
					if buffer[position] != rune(':') {
						goto l133
					}
					position++
				}
			l135:
				depth--
				add(ruleScenarioKeyWord, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 2 StepKeyWord <- <((&('*') '*') | (&('B') ('B' 'u' 't')) | (&('O') ('O' 'r')) | (&('A') ('A' 'n' 'd')) | (&('T') ('T' 'h' 'e' 'n')) | (&('W') ('W' 'h' 'e' 'n')) | (&('G') ('G' 'i' 'v' 'e' 'n')))> */
		func() bool {
			position138, tokenIndex138, depth138 := position, tokenIndex, depth
			{
				position139 := position
				depth++
				{
					switch buffer[position] {
					case '*':
						if buffer[position] != rune('*') {
							goto l138
						}
						position++
						break
					case 'B':
						if buffer[position] != rune('B') {
							goto l138
						}
						position++
						if buffer[position] != rune('u') {
							goto l138
						}
						position++
						if buffer[position] != rune('t') {
							goto l138
						}
						position++
						break
					case 'O':
						if buffer[position] != rune('O') {
							goto l138
						}
						position++
						if buffer[position] != rune('r') {
							goto l138
						}
						position++
						break
					case 'A':
						if buffer[position] != rune('A') {
							goto l138
						}
						position++
						if buffer[position] != rune('n') {
							goto l138
						}
						position++
						if buffer[position] != rune('d') {
							goto l138
						}
						position++
						break
					case 'T':
						if buffer[position] != rune('T') {
							goto l138
						}
						position++
						if buffer[position] != rune('h') {
							goto l138
						}
						position++
						if buffer[position] != rune('e') {
							goto l138
						}
						position++
						if buffer[position] != rune('n') {
							goto l138
						}
						position++
						break
					case 'W':
						if buffer[position] != rune('W') {
							goto l138
						}
						position++
						if buffer[position] != rune('h') {
							goto l138
						}
						position++
						if buffer[position] != rune('e') {
							goto l138
						}
						position++
						if buffer[position] != rune('n') {
							goto l138
						}
						position++
						break
					default:
						if buffer[position] != rune('G') {
							goto l138
						}
						position++
						if buffer[position] != rune('i') {
							goto l138
						}
						position++
						if buffer[position] != rune('v') {
							goto l138
						}
						position++
						if buffer[position] != rune('e') {
							goto l138
						}
						position++
						if buffer[position] != rune('n') {
							goto l138
						}
						position++
						break
					}
				}

				depth--
				add(ruleStepKeyWord, position139)
			}
			return true
		l138:
			position, tokenIndex, depth = position138, tokenIndex138, depth138
			return false
		},
		/* 3 Feature <- <(Tags ('F' 'e' 'a' 't' 'u' 'r' 'e' ':') WS* <UntilLineEnd?> Action0 <> Action1 LineEnd (WS* !(('@' Word) / ScenarioKeyWord) <UntilLineEnd?> Action2 LineEnd Action3)* Action4 (Background / Scenario / Outline / BlankLine)* Action5)> */
		nil,
		/* 4 Background <- <(Tags ('B' 'a' 'c' 'k' 'g' 'r' 'o' 'u' 'n' 'd' ':') WS* <UntilLineEnd?> Action6 <> Action7 LineEnd (WS* !(('@' Word) / ScenarioKeyWord / StepKeyWord) <UntilLineEnd?> Action8 LineEnd Action9)* Action10 (Step / BlankLine)* Action11)> */
		nil,
		/* 5 Scenario <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ':') WS* <UntilLineEnd?> Action12 <> Action13 LineEnd (WS* !(('@' Word) / ScenarioKeyWord / StepKeyWord) <UntilLineEnd?> Action14 LineEnd Action15)* Action16 (Step / BlankLine)* Action17)> */
		nil,
		/* 6 Outline <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':') WS* <UntilLineEnd?> Action18 <> Action19 LineEnd (WS* !(('@' Word) / ScenarioKeyWord / StepKeyWord) <UntilLineEnd?> Action20 LineEnd Action21)* Action22 (Step / BlankLine)* (OutlineExamples / BlankLine)* Action23)> */
		nil,
		/* 7 OutlineExamples <- <(OS ('E' 'x' 'a' 'm' 'p' 'l' 'e' 's' ':') WS* <UntilLineEnd?> Action24 LineEnd Action25 Table? Action26)> */
		nil,
		/* 8 Step <- <(WS* <StepKeyWord> Action27 WS* <UntilLineEnd> Action28 LineEnd Action29 StepArgument? Action30)> */
		func() bool {
			position146, tokenIndex146, depth146 := position, tokenIndex, depth
			{
				position147 := position
				depth++
			l148:
				{
					position149, tokenIndex149, depth149 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l149
					}
					goto l148
				l149:
					position, tokenIndex, depth = position149, tokenIndex149, depth149
				}
				{
					position150 := position
					depth++
					if !_rules[ruleStepKeyWord]() {
						goto l146
					}
					depth--
					add(rulePegText, position150)
				}
				{
					add(ruleAction27, position)
				}
			l152:
				{
					position153, tokenIndex153, depth153 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l153
					}
					goto l152
				l153:
					position, tokenIndex, depth = position153, tokenIndex153, depth153
				}
				{
					position154 := position
					depth++
					if !_rules[ruleUntilLineEnd]() {
						goto l146
					}
					depth--
					add(rulePegText, position154)
				}
				{
					add(ruleAction28, position)
				}
				if !_rules[ruleLineEnd]() {
					goto l146
				}
				{
					add(ruleAction29, position)
				}
				{
					position157, tokenIndex157, depth157 := position, tokenIndex, depth
					{
						position159 := position
						depth++
						{
							position160, tokenIndex160, depth160 := position, tokenIndex, depth
							if !_rules[ruleTable]() {
								goto l161
							}
							goto l160
						l161:
							position, tokenIndex, depth = position160, tokenIndex160, depth160
							{
								position162 := position
								depth++
							l163:
								{
									position164, tokenIndex164, depth164 := position, tokenIndex, depth
								l165:
									{
										position166, tokenIndex166, depth166 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l166
										}
										goto l165
									l166:
										position, tokenIndex, depth = position166, tokenIndex166, depth166
									}
									if !_rules[ruleNL]() {
										goto l164
									}
									goto l163
								l164:
									position, tokenIndex, depth = position164, tokenIndex164, depth164
								}
								{
									position167 := position
									depth++
								l168:
									{
										position169, tokenIndex169, depth169 := position, tokenIndex, depth
										if !_rules[ruleWS]() {
											goto l169
										}
										goto l168
									l169:
										position, tokenIndex, depth = position169, tokenIndex169, depth169
									}
									depth--
									add(rulePegText, position167)
								}
								if !_rules[rulePyStringQuote]() {
									goto l157
								}
								if !_rules[ruleNL]() {
									goto l157
								}
								{
									add(ruleAction31, position)
								}
							l171:
								{
									position172, tokenIndex172, depth172 := position, tokenIndex, depth
									{
										position173, tokenIndex173, depth173 := position, tokenIndex, depth
									l174:
										{
											position175, tokenIndex175, depth175 := position, tokenIndex, depth
											if !_rules[ruleWS]() {
												goto l175
											}
											goto l174
										l175:
											position, tokenIndex, depth = position175, tokenIndex175, depth175
										}
										if !_rules[rulePyStringQuote]() {
											goto l173
										}
										goto l172
									l173:
										position, tokenIndex, depth = position173, tokenIndex173, depth173
									}
									{
										position176 := position
										depth++
										{
											position177 := position
											depth++
											{
												position178 := position
												depth++
											l179:
												{
													position180, tokenIndex180, depth180 := position, tokenIndex, depth
													{
														position181, tokenIndex181, depth181 := position, tokenIndex, depth
														if buffer[position] != rune('\n') {
															goto l181
														}
														position++
														goto l180
													l181:
														position, tokenIndex, depth = position181, tokenIndex181, depth181
													}
													if !matchDot() {
														goto l180
													}
													goto l179
												l180:
													position, tokenIndex, depth = position180, tokenIndex180, depth180
												}
												depth--
												add(ruleUntilNL, position178)
											}
											depth--
											add(rulePegText, position177)
										}
										if !_rules[ruleNL]() {
											goto l172
										}
										{
											add(ruleAction33, position)
										}
										depth--
										add(rulePyStringLine, position176)
									}
									goto l171
								l172:
									position, tokenIndex, depth = position172, tokenIndex172, depth172
								}
							l183:
								{
									position184, tokenIndex184, depth184 := position, tokenIndex, depth
									if !_rules[ruleWS]() {
										goto l184
									}
									goto l183
								l184:
									position, tokenIndex, depth = position184, tokenIndex184, depth184
								}
								if !_rules[rulePyStringQuote]() {
									goto l157
								}
								if !_rules[ruleLineEnd]() {
									goto l157
								}
								{
									add(ruleAction32, position)
								}
								depth--
								add(rulePyString, position162)
							}
						}
					l160:
						depth--
						add(ruleStepArgument, position159)
					}
					goto l158
				l157:
					position, tokenIndex, depth = position157, tokenIndex157, depth157
				}
			l158:
				{
					add(ruleAction30, position)
				}
				depth--
				add(ruleStep, position147)
			}
			return true
		l146:
			position, tokenIndex, depth = position146, tokenIndex146, depth146
			return false
		},
		/* 9 StepArgument <- <(Table / PyString)> */
		nil,
		/* 10 PyString <- <((WS* NL)* <WS*> PyStringQuote NL Action31 (!(WS* PyStringQuote) PyStringLine)* WS* PyStringQuote LineEnd Action32)> */
		nil,
		/* 11 PyStringQuote <- <('"' '"' '"')> */
		func() bool {
			position189, tokenIndex189, depth189 := position, tokenIndex, depth
			{
				position190 := position
				depth++
				if buffer[position] != rune('"') {
					goto l189
				}
				position++
				if buffer[position] != rune('"') {
					goto l189
				}
				position++
				if buffer[position] != rune('"') {
					goto l189
				}
				position++
				depth--
				add(rulePyStringQuote, position190)
			}
			return true
		l189:
			position, tokenIndex, depth = position189, tokenIndex189, depth189
			return false
		},
		/* 12 PyStringLine <- <(<UntilNL> NL Action33)> */
		nil,
		/* 13 Table <- <(Action34 TableRow+ Action35)> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				{
					add(ruleAction34, position)
				}
				{
					position197 := position
					depth++
					{
						add(ruleAction36, position)
					}
					if !_rules[ruleOS]() {
						goto l192
					}
					if buffer[position] != rune('|') {
						goto l192
					}
					position++
					{
						position201 := position
						depth++
						{
							position202 := position
							depth++
							{
								position205, tokenIndex205, depth205 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '|':
										if buffer[position] != rune('|') {
											goto l205
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l205
										}
										position++
										break
									default:
										if buffer[position] != rune('\r') {
											goto l205
										}
										position++
										break
									}
								}

								goto l192
							l205:
								position, tokenIndex, depth = position205, tokenIndex205, depth205
							}
							if !matchDot() {
								goto l192
							}
						l203:
							{
								position204, tokenIndex204, depth204 := position, tokenIndex, depth
								{
									position207, tokenIndex207, depth207 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l207
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l207
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l207
											}
											position++
											break
										}
									}

									goto l204
								l207:
									position, tokenIndex, depth = position207, tokenIndex207, depth207
								}
								if !matchDot() {
									goto l204
								}
								goto l203
							l204:
								position, tokenIndex, depth = position204, tokenIndex204, depth204
							}
							depth--
							add(rulePegText, position202)
						}
						if buffer[position] != rune('|') {
							goto l192
						}
						position++
						{
							add(ruleAction38, position)
						}
						depth--
						add(ruleTableCell, position201)
					}
				l199:
					{
						position200, tokenIndex200, depth200 := position, tokenIndex, depth
						{
							position210 := position
							depth++
							{
								position211 := position
								depth++
								{
									position214, tokenIndex214, depth214 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l214
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l214
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l214
											}
											position++
											break
										}
									}

									goto l200
								l214:
									position, tokenIndex, depth = position214, tokenIndex214, depth214
								}
								if !matchDot() {
									goto l200
								}
							l212:
								{
									position213, tokenIndex213, depth213 := position, tokenIndex, depth
									{
										position216, tokenIndex216, depth216 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
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

										goto l213
									l216:
										position, tokenIndex, depth = position216, tokenIndex216, depth216
									}
									if !matchDot() {
										goto l213
									}
									goto l212
								l213:
									position, tokenIndex, depth = position213, tokenIndex213, depth213
								}
								depth--
								add(rulePegText, position211)
							}
							if buffer[position] != rune('|') {
								goto l200
							}
							position++
							{
								add(ruleAction38, position)
							}
							depth--
							add(ruleTableCell, position210)
						}
						goto l199
					l200:
						position, tokenIndex, depth = position200, tokenIndex200, depth200
					}
					if !_rules[ruleLineEnd]() {
						goto l192
					}
					{
						add(ruleAction37, position)
					}
					depth--
					add(ruleTableRow, position197)
				}
			l195:
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					{
						position220 := position
						depth++
						{
							add(ruleAction36, position)
						}
						if !_rules[ruleOS]() {
							goto l196
						}
						if buffer[position] != rune('|') {
							goto l196
						}
						position++
						{
							position224 := position
							depth++
							{
								position225 := position
								depth++
								{
									position228, tokenIndex228, depth228 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l228
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l228
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l228
											}
											position++
											break
										}
									}

									goto l196
								l228:
									position, tokenIndex, depth = position228, tokenIndex228, depth228
								}
								if !matchDot() {
									goto l196
								}
							l226:
								{
									position227, tokenIndex227, depth227 := position, tokenIndex, depth
									{
										position230, tokenIndex230, depth230 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l230
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l230
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l230
												}
												position++
												break
											}
										}

										goto l227
									l230:
										position, tokenIndex, depth = position230, tokenIndex230, depth230
									}
									if !matchDot() {
										goto l227
									}
									goto l226
								l227:
									position, tokenIndex, depth = position227, tokenIndex227, depth227
								}
								depth--
								add(rulePegText, position225)
							}
							if buffer[position] != rune('|') {
								goto l196
							}
							position++
							{
								add(ruleAction38, position)
							}
							depth--
							add(ruleTableCell, position224)
						}
					l222:
						{
							position223, tokenIndex223, depth223 := position, tokenIndex, depth
							{
								position233 := position
								depth++
								{
									position234 := position
									depth++
									{
										position237, tokenIndex237, depth237 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l237
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l237
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l237
												}
												position++
												break
											}
										}

										goto l223
									l237:
										position, tokenIndex, depth = position237, tokenIndex237, depth237
									}
									if !matchDot() {
										goto l223
									}
								l235:
									{
										position236, tokenIndex236, depth236 := position, tokenIndex, depth
										{
											position239, tokenIndex239, depth239 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '|':
													if buffer[position] != rune('|') {
														goto l239
													}
													position++
													break
												case '\n':
													if buffer[position] != rune('\n') {
														goto l239
													}
													position++
													break
												default:
													if buffer[position] != rune('\r') {
														goto l239
													}
													position++
													break
												}
											}

											goto l236
										l239:
											position, tokenIndex, depth = position239, tokenIndex239, depth239
										}
										if !matchDot() {
											goto l236
										}
										goto l235
									l236:
										position, tokenIndex, depth = position236, tokenIndex236, depth236
									}
									depth--
									add(rulePegText, position234)
								}
								if buffer[position] != rune('|') {
									goto l223
								}
								position++
								{
									add(ruleAction38, position)
								}
								depth--
								add(ruleTableCell, position233)
							}
							goto l222
						l223:
							position, tokenIndex, depth = position223, tokenIndex223, depth223
						}
						if !_rules[ruleLineEnd]() {
							goto l196
						}
						{
							add(ruleAction37, position)
						}
						depth--
						add(ruleTableRow, position220)
					}
					goto l195
				l196:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
				}
				{
					add(ruleAction35, position)
				}
				depth--
				add(ruleTable, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 14 TableRow <- <(Action36 OS '|' TableCell+ LineEnd Action37)> */
		nil,
		/* 15 TableCell <- <(<(!((&('|') '|') | (&('\n') '\n') | (&('\r') '\r')) .)+> '|' Action38)> */
		nil,
		/* 16 Tags <- <((Tag+ WS* LineEnd?)* OS)> */
		func() bool {
			position246, tokenIndex246, depth246 := position, tokenIndex, depth
			{
				position247 := position
				depth++
			l248:
				{
					position249, tokenIndex249, depth249 := position, tokenIndex, depth
					{
						position252 := position
						depth++
						if !_rules[ruleOS]() {
							goto l249
						}
						if buffer[position] != rune('@') {
							goto l249
						}
						position++
						{
							position253 := position
							depth++
							if !_rules[ruleWord]() {
								goto l249
							}
							depth--
							add(rulePegText, position253)
						}
						{
							add(ruleAction39, position)
						}
						depth--
						add(ruleTag, position252)
					}
				l250:
					{
						position251, tokenIndex251, depth251 := position, tokenIndex, depth
						{
							position255 := position
							depth++
							if !_rules[ruleOS]() {
								goto l251
							}
							if buffer[position] != rune('@') {
								goto l251
							}
							position++
							{
								position256 := position
								depth++
								if !_rules[ruleWord]() {
									goto l251
								}
								depth--
								add(rulePegText, position256)
							}
							{
								add(ruleAction39, position)
							}
							depth--
							add(ruleTag, position255)
						}
						goto l250
					l251:
						position, tokenIndex, depth = position251, tokenIndex251, depth251
					}
				l258:
					{
						position259, tokenIndex259, depth259 := position, tokenIndex, depth
						if !_rules[ruleWS]() {
							goto l259
						}
						goto l258
					l259:
						position, tokenIndex, depth = position259, tokenIndex259, depth259
					}
					{
						position260, tokenIndex260, depth260 := position, tokenIndex, depth
						if !_rules[ruleLineEnd]() {
							goto l260
						}
						goto l261
					l260:
						position, tokenIndex, depth = position260, tokenIndex260, depth260
					}
				l261:
					goto l248
				l249:
					position, tokenIndex, depth = position249, tokenIndex249, depth249
				}
				if !_rules[ruleOS]() {
					goto l246
				}
				depth--
				add(ruleTags, position247)
			}
			return true
		l246:
			position, tokenIndex, depth = position246, tokenIndex246, depth246
			return false
		},
		/* 17 Tag <- <(OS '@' <Word> Action39)> */
		nil,
		/* 18 Word <- <(!((&('#') '#') | (&('"') '"') | (&(' ') ' ') | (&('\t') '\t') | (&('\n') '\n') | (&('\r') '\r')) .)+> */
		func() bool {
			position263, tokenIndex263, depth263 := position, tokenIndex, depth
			{
				position264 := position
				depth++
				{
					position267, tokenIndex267, depth267 := position, tokenIndex, depth
					{
						switch buffer[position] {
						case '#':
							if buffer[position] != rune('#') {
								goto l267
							}
							position++
							break
						case '"':
							if buffer[position] != rune('"') {
								goto l267
							}
							position++
							break
						case ' ':
							if buffer[position] != rune(' ') {
								goto l267
							}
							position++
							break
						case '\t':
							if buffer[position] != rune('\t') {
								goto l267
							}
							position++
							break
						case '\n':
							if buffer[position] != rune('\n') {
								goto l267
							}
							position++
							break
						default:
							if buffer[position] != rune('\r') {
								goto l267
							}
							position++
							break
						}
					}

					goto l263
				l267:
					position, tokenIndex, depth = position267, tokenIndex267, depth267
				}
				if !matchDot() {
					goto l263
				}
			l265:
				{
					position266, tokenIndex266, depth266 := position, tokenIndex, depth
					{
						position269, tokenIndex269, depth269 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '#':
								if buffer[position] != rune('#') {
									goto l269
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l269
								}
								position++
								break
							case ' ':
								if buffer[position] != rune(' ') {
									goto l269
								}
								position++
								break
							case '\t':
								if buffer[position] != rune('\t') {
									goto l269
								}
								position++
								break
							case '\n':
								if buffer[position] != rune('\n') {
									goto l269
								}
								position++
								break
							default:
								if buffer[position] != rune('\r') {
									goto l269
								}
								position++
								break
							}
						}

						goto l266
					l269:
						position, tokenIndex, depth = position269, tokenIndex269, depth269
					}
					if !matchDot() {
						goto l266
					}
					goto l265
				l266:
					position, tokenIndex, depth = position266, tokenIndex266, depth266
				}
				depth--
				add(ruleWord, position264)
			}
			return true
		l263:
			position, tokenIndex, depth = position263, tokenIndex263, depth263
			return false
		},
		/* 19 EscapedChar <- <('\\' .)> */
		func() bool {
			position271, tokenIndex271, depth271 := position, tokenIndex, depth
			{
				position272 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l271
				}
				position++
				if !matchDot() {
					goto l271
				}
				depth--
				add(ruleEscapedChar, position272)
			}
			return true
		l271:
			position, tokenIndex, depth = position271, tokenIndex271, depth271
			return false
		},
		/* 20 QuotedString <- <('"' (EscapedChar / (!((&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+)* '"')> */
		nil,
		/* 21 UntilLineEnd <- <(EscapedChar / (!((&('#') '#') | (&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+ / QuotedString)+> */
		func() bool {
			position274, tokenIndex274, depth274 := position, tokenIndex, depth
			{
				position275 := position
				depth++
				{
					position278, tokenIndex278, depth278 := position, tokenIndex, depth
					if !_rules[ruleEscapedChar]() {
						goto l279
					}
					goto l278
				l279:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					{
						position283, tokenIndex283, depth283 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '#':
								if buffer[position] != rune('#') {
									goto l283
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l283
								}
								position++
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l283
								}
								position++
								break
							default:
								if buffer[position] != rune('\n') {
									goto l283
								}
								position++
								break
							}
						}

						goto l280
					l283:
						position, tokenIndex, depth = position283, tokenIndex283, depth283
					}
					if !matchDot() {
						goto l280
					}
				l281:
					{
						position282, tokenIndex282, depth282 := position, tokenIndex, depth
						{
							position285, tokenIndex285, depth285 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l285
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l285
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l285
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l285
									}
									position++
									break
								}
							}

							goto l282
						l285:
							position, tokenIndex, depth = position285, tokenIndex285, depth285
						}
						if !matchDot() {
							goto l282
						}
						goto l281
					l282:
						position, tokenIndex, depth = position282, tokenIndex282, depth282
					}
					goto l278
				l280:
					position, tokenIndex, depth = position278, tokenIndex278, depth278
					{
						position287 := position
						depth++
						if buffer[position] != rune('"') {
							goto l274
						}
						position++
					l288:
						{
							position289, tokenIndex289, depth289 := position, tokenIndex, depth
							{
								position290, tokenIndex290, depth290 := position, tokenIndex, depth
								if !_rules[ruleEscapedChar]() {
									goto l291
								}
								goto l290
							l291:
								position, tokenIndex, depth = position290, tokenIndex290, depth290
								{
									position294, tokenIndex294, depth294 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '"':
											if buffer[position] != rune('"') {
												goto l294
											}
											position++
											break
										case '\\':
											if buffer[position] != rune('\\') {
												goto l294
											}
											position++
											break
										default:
											if buffer[position] != rune('\n') {
												goto l294
											}
											position++
											break
										}
									}

									goto l289
								l294:
									position, tokenIndex, depth = position294, tokenIndex294, depth294
								}
								if !matchDot() {
									goto l289
								}
							l292:
								{
									position293, tokenIndex293, depth293 := position, tokenIndex, depth
									{
										position296, tokenIndex296, depth296 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l296
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l296
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l296
												}
												position++
												break
											}
										}

										goto l293
									l296:
										position, tokenIndex, depth = position296, tokenIndex296, depth296
									}
									if !matchDot() {
										goto l293
									}
									goto l292
								l293:
									position, tokenIndex, depth = position293, tokenIndex293, depth293
								}
							}
						l290:
							goto l288
						l289:
							position, tokenIndex, depth = position289, tokenIndex289, depth289
						}
						if buffer[position] != rune('"') {
							goto l274
						}
						position++
						depth--
						add(ruleQuotedString, position287)
					}
				}
			l278:
			l276:
				{
					position277, tokenIndex277, depth277 := position, tokenIndex, depth
					{
						position298, tokenIndex298, depth298 := position, tokenIndex, depth
						if !_rules[ruleEscapedChar]() {
							goto l299
						}
						goto l298
					l299:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
						{
							position303, tokenIndex303, depth303 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l303
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l303
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l303
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l303
									}
									position++
									break
								}
							}

							goto l300
						l303:
							position, tokenIndex, depth = position303, tokenIndex303, depth303
						}
						if !matchDot() {
							goto l300
						}
					l301:
						{
							position302, tokenIndex302, depth302 := position, tokenIndex, depth
							{
								position305, tokenIndex305, depth305 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '#':
										if buffer[position] != rune('#') {
											goto l305
										}
										position++
										break
									case '"':
										if buffer[position] != rune('"') {
											goto l305
										}
										position++
										break
									case '\\':
										if buffer[position] != rune('\\') {
											goto l305
										}
										position++
										break
									default:
										if buffer[position] != rune('\n') {
											goto l305
										}
										position++
										break
									}
								}

								goto l302
							l305:
								position, tokenIndex, depth = position305, tokenIndex305, depth305
							}
							if !matchDot() {
								goto l302
							}
							goto l301
						l302:
							position, tokenIndex, depth = position302, tokenIndex302, depth302
						}
						goto l298
					l300:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
						{
							position307 := position
							depth++
							if buffer[position] != rune('"') {
								goto l277
							}
							position++
						l308:
							{
								position309, tokenIndex309, depth309 := position, tokenIndex, depth
								{
									position310, tokenIndex310, depth310 := position, tokenIndex, depth
									if !_rules[ruleEscapedChar]() {
										goto l311
									}
									goto l310
								l311:
									position, tokenIndex, depth = position310, tokenIndex310, depth310
									{
										position314, tokenIndex314, depth314 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l314
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l314
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l314
												}
												position++
												break
											}
										}

										goto l309
									l314:
										position, tokenIndex, depth = position314, tokenIndex314, depth314
									}
									if !matchDot() {
										goto l309
									}
								l312:
									{
										position313, tokenIndex313, depth313 := position, tokenIndex, depth
										{
											position316, tokenIndex316, depth316 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '"':
													if buffer[position] != rune('"') {
														goto l316
													}
													position++
													break
												case '\\':
													if buffer[position] != rune('\\') {
														goto l316
													}
													position++
													break
												default:
													if buffer[position] != rune('\n') {
														goto l316
													}
													position++
													break
												}
											}

											goto l313
										l316:
											position, tokenIndex, depth = position316, tokenIndex316, depth316
										}
										if !matchDot() {
											goto l313
										}
										goto l312
									l313:
										position, tokenIndex, depth = position313, tokenIndex313, depth313
									}
								}
							l310:
								goto l308
							l309:
								position, tokenIndex, depth = position309, tokenIndex309, depth309
							}
							if buffer[position] != rune('"') {
								goto l277
							}
							position++
							depth--
							add(ruleQuotedString, position307)
						}
					}
				l298:
					goto l276
				l277:
					position, tokenIndex, depth = position277, tokenIndex277, depth277
				}
				depth--
				add(ruleUntilLineEnd, position275)
			}
			return true
		l274:
			position, tokenIndex, depth = position274, tokenIndex274, depth274
			return false
		},
		/* 22 LineEnd <- <(WS* LineComment? NL)> */
		func() bool {
			position318, tokenIndex318, depth318 := position, tokenIndex, depth
			{
				position319 := position
				depth++
			l320:
				{
					position321, tokenIndex321, depth321 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l321
					}
					goto l320
				l321:
					position, tokenIndex, depth = position321, tokenIndex321, depth321
				}
				{
					position322, tokenIndex322, depth322 := position, tokenIndex, depth
					if !_rules[ruleLineComment]() {
						goto l322
					}
					goto l323
				l322:
					position, tokenIndex, depth = position322, tokenIndex322, depth322
				}
			l323:
				if !_rules[ruleNL]() {
					goto l318
				}
				depth--
				add(ruleLineEnd, position319)
			}
			return true
		l318:
			position, tokenIndex, depth = position318, tokenIndex318, depth318
			return false
		},
		/* 23 LineComment <- <('#' <(!'\n' .)*> Action40)> */
		func() bool {
			position324, tokenIndex324, depth324 := position, tokenIndex, depth
			{
				position325 := position
				depth++
				if buffer[position] != rune('#') {
					goto l324
				}
				position++
				{
					position326 := position
					depth++
				l327:
					{
						position328, tokenIndex328, depth328 := position, tokenIndex, depth
						{
							position329, tokenIndex329, depth329 := position, tokenIndex, depth
							if buffer[position] != rune('\n') {
								goto l329
							}
							position++
							goto l328
						l329:
							position, tokenIndex, depth = position329, tokenIndex329, depth329
						}
						if !matchDot() {
							goto l328
						}
						goto l327
					l328:
						position, tokenIndex, depth = position328, tokenIndex328, depth328
					}
					depth--
					add(rulePegText, position326)
				}
				{
					add(ruleAction40, position)
				}
				depth--
				add(ruleLineComment, position325)
			}
			return true
		l324:
			position, tokenIndex, depth = position324, tokenIndex324, depth324
			return false
		},
		/* 24 BlankLine <- <(((WS LineEnd) / (LineComment? NL)) Action41)> */
		func() bool {
			position331, tokenIndex331, depth331 := position, tokenIndex, depth
			{
				position332 := position
				depth++
				{
					position333, tokenIndex333, depth333 := position, tokenIndex, depth
					if !_rules[ruleWS]() {
						goto l334
					}
					if !_rules[ruleLineEnd]() {
						goto l334
					}
					goto l333
				l334:
					position, tokenIndex, depth = position333, tokenIndex333, depth333
					{
						position335, tokenIndex335, depth335 := position, tokenIndex, depth
						if !_rules[ruleLineComment]() {
							goto l335
						}
						goto l336
					l335:
						position, tokenIndex, depth = position335, tokenIndex335, depth335
					}
				l336:
					if !_rules[ruleNL]() {
						goto l331
					}
				}
			l333:
				{
					add(ruleAction41, position)
				}
				depth--
				add(ruleBlankLine, position332)
			}
			return true
		l331:
			position, tokenIndex, depth = position331, tokenIndex331, depth331
			return false
		},
		/* 25 OS <- <(NL / WS)*> */
		func() bool {
			{
				position339 := position
				depth++
			l340:
				{
					position341, tokenIndex341, depth341 := position, tokenIndex, depth
					{
						position342, tokenIndex342, depth342 := position, tokenIndex, depth
						if !_rules[ruleNL]() {
							goto l343
						}
						goto l342
					l343:
						position, tokenIndex, depth = position342, tokenIndex342, depth342
						if !_rules[ruleWS]() {
							goto l341
						}
					}
				l342:
					goto l340
				l341:
					position, tokenIndex, depth = position341, tokenIndex341, depth341
				}
				depth--
				add(ruleOS, position339)
			}
			return true
		},
		/* 26 WS <- <(' ' / '\t')> */
		func() bool {
			position344, tokenIndex344, depth344 := position, tokenIndex, depth
			{
				position345 := position
				depth++
				{
					position346, tokenIndex346, depth346 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l347
					}
					position++
					goto l346
				l347:
					position, tokenIndex, depth = position346, tokenIndex346, depth346
					if buffer[position] != rune('\t') {
						goto l344
					}
					position++
				}
			l346:
				depth--
				add(ruleWS, position345)
			}
			return true
		l344:
			position, tokenIndex, depth = position344, tokenIndex344, depth344
			return false
		},
		/* 27 UntilNL <- <(!'\n' .)*> */
		nil,
		/* 28 NL <- <('\n' / '\r' / ('\r' '\n'))> */
		func() bool {
			position349, tokenIndex349, depth349 := position, tokenIndex, depth
			{
				position350 := position
				depth++
				{
					position351, tokenIndex351, depth351 := position, tokenIndex, depth
					if buffer[position] != rune('\n') {
						goto l352
					}
					position++
					goto l351
				l352:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
					if buffer[position] != rune('\r') {
						goto l353
					}
					position++
					goto l351
				l353:
					position, tokenIndex, depth = position351, tokenIndex351, depth351
					if buffer[position] != rune('\r') {
						goto l349
					}
					position++
					if buffer[position] != rune('\n') {
						goto l349
					}
					position++
				}
			l351:
				depth--
				add(ruleNL, position350)
			}
			return true
		l349:
			position, tokenIndex, depth = position349, tokenIndex349, depth349
			return false
		},
		nil,
		/* 31 Action0 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 32 Action1 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 33 Action2 <- <{ p.buf2 = p.buf2 + buffer[begin:end] }> */
		nil,
		/* 34 Action3 <- <{ p.buf2 = p.buf2 + "\n" }> */
		nil,
		/* 35 Action4 <- <{ p.beginFeature(trimWS(p.buf1), trimWSML(p.buf2), p.buftags); p.buftags = nil }> */
		nil,
		/* 36 Action5 <- <{ p.endFeature() }> */
		nil,
		/* 37 Action6 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 38 Action7 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 39 Action8 <- <{ p.buf2 = p.buf2 + buffer[begin:end] }> */
		nil,
		/* 40 Action9 <- <{ p.buf2 = p.buf2 + "\n" }> */
		nil,
		/* 41 Action10 <- <{ p.beginBackground(trimWS(p.buf1), trimWSML(p.buf2), p.buftags); p.buftags = nil }> */
		nil,
		/* 42 Action11 <- <{ p.endBackground() }> */
		nil,
		/* 43 Action12 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 44 Action13 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 45 Action14 <- <{ p.buf2 = p.buf2 + buffer[begin:end] }> */
		nil,
		/* 46 Action15 <- <{ p.buf2 = p.buf2 + "\n" }> */
		nil,
		/* 47 Action16 <- <{ p.beginScenario(trimWS(p.buf1), trimWSML(p.buf2), p.buftags); p.buftags = nil }> */
		nil,
		/* 48 Action17 <- <{ p.endScenario() }> */
		nil,
		/* 49 Action18 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 50 Action19 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 51 Action20 <- <{ p.buf2 = p.buf2 + buffer[begin:end] }> */
		nil,
		/* 52 Action21 <- <{ p.buf2 = p.buf2 + "\n" }> */
		nil,
		/* 53 Action22 <- <{ p.beginOutline(trimWS(p.buf1), trimWSML(p.buf2), p.buftags); p.buftags = nil }> */
		nil,
		/* 54 Action23 <- <{ p.endOutline() }> */
		nil,
		/* 55 Action24 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 56 Action25 <- <{ p.beginOutlineExamples(trimWS(p.buf1)) }> */
		nil,
		/* 57 Action26 <- <{ p.endOutlineExamples() }> */
		nil,
		/* 58 Action27 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 59 Action28 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 60 Action29 <- <{ p.beginStep(trimWS(p.buf1), trimWS(p.buf2)) }> */
		nil,
		/* 61 Action30 <- <{ p.endStep() }> */
		nil,
		/* 62 Action31 <- <{ p.beginPyString(buffer[begin:end]) }> */
		nil,
		/* 63 Action32 <- <{ p.endPyString() }> */
		nil,
		/* 64 Action33 <- <{ p.bufferPyString(buffer[begin:end]) }> */
		nil,
		/* 65 Action34 <- <{ p.beginTable() }> */
		nil,
		/* 66 Action35 <- <{ p.endTable() }> */
		nil,
		/* 67 Action36 <- <{ p.beginTableRow() }> */
		nil,
		/* 68 Action37 <- <{ p.endTableRow() }> */
		nil,
		/* 69 Action38 <- <{ p.beginTableCell(); p.endTableCell(trimWS(buffer[begin:end])) }> */
		nil,
		/* 70 Action39 <- <{ p.buftags = append(p.buftags, buffer[begin:end]) }> */
		nil,
		/* 71 Action40 <- <{ p.bufcmt = buffer[begin:end]; p.triggerComment(p.bufcmt) }> */
		nil,
		/* 72 Action41 <- <{ p.triggerBlankLine() }> */
		nil,
	}
	p.rules = _rules
}
