package gherkin

import (
	/*"bytes"*/
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type rule uint8

const (
	ruleUnknown rule = iota
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
	ruleStringToEndOfLine
	ruleOS
	ruleWS
	ruleNL
	ruleWSNLEOF
	ruleMLWS
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
	"StringToEndOfLine",
	"OS",
	"WS",
	"NL",
	"WSNLEOF",
	"MLWS",
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

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule rule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	Error() []token32
	trim(length int)
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	rule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.rule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) GetToken32() token32 {
	return token32{rule: t.rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.rule], t.begin, t.end, t.next)
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
		if token.rule == ruleUnknown {
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
			state, S.rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.rule, t.begin, t.end, int16(depth), leaf
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
							write(token16{rule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{rule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.rule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.rule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{rule: rule_Suf, begin: b.end, end: a.end}, true)
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
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.rule])
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
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.rule])
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
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule rule, begin, end, depth, index int) {
	t.tree[index] = token16{rule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
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
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	rule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.rule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) GetToken32() token32 {
	return token32{rule: t.rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.rule], t.begin, t.end, t.next)
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
		if token.rule == ruleUnknown {
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
			state, S.rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.rule, t.begin, t.end, int32(depth), leaf
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
							write(token32{rule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{rule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.rule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.rule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{rule: rule_Suf, begin: b.end, end: a.end}, true)
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
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.rule])
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
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.rule])
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
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule rule, begin, end, depth, index int) {
	t.tree[index] = token32{rule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
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
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.GetToken32()
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

	buffer2   string
	tagBuffer []string

	Buffer string
	buffer []rune
	rules  [47]func() bool
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
			rul3s[token.rule],
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
		switch token.rule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruleAction0:
			p.buffer2 = buffer[begin:end]
		case ruleAction1:
			p.beginFeature(p.buffer2, buffer[begin:end], p.tagBuffer)
			p.tagBuffer = nil
		case ruleAction2:
			p.endFeature()
		case ruleAction3:
			p.beginBackground(buffer[begin:end], p.tagBuffer)
			p.tagBuffer = nil
		case ruleAction4:
			p.endBackground()
		case ruleAction5:
			p.beginScenario(buffer[begin:end], p.tagBuffer)
			p.tagBuffer = nil
		case ruleAction6:
			p.endScenario()
		case ruleAction7:
			p.beginOutline(buffer[begin:end], p.tagBuffer)
			p.tagBuffer = nil
		case ruleAction8:
			p.endOutline()
		case ruleAction9:
			p.beginOutlineExamples()
		case ruleAction10:
			p.endOutlineExamples()
		case ruleAction11:
			p.buffer2 = buffer[begin:end]
		case ruleAction12:
			p.beginStep(p.buffer2, buffer[begin:end])
		case ruleAction13:
			p.endStep()
		case ruleAction14:
			p.beginPyString(buffer[begin:end])
		case ruleAction15:
			p.endPyString()
		case ruleAction16:
			p.bufferPyString(buffer[begin:end])
		case ruleAction17:
			p.beginTable()
		case ruleAction18:
			p.endTable()
		case ruleAction19:
			p.beginTableRow()
		case ruleAction20:
			p.endTableRow()
		case ruleAction21:
			p.beginTableCell()
			p.endTableCell(buffer[begin:end])
		case ruleAction22:
			p.tagBuffer = append(p.tagBuffer, buffer[begin:end])

		}
	}
}

func (p *gherkinPeg) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.buffer, p.rules

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

	add := func(rule rule, begin int) {
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

	rules = [...]func() bool{
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
						if !rules[ruleTags]() {
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
							if !rules[ruleWS]() {
								goto l6
							}
							goto l5
						l6:
							position, tokenIndex, depth = position6, tokenIndex6, depth6
						}
						{
							position7 := position
							depth++
							if !rules[ruleStringToEndOfLine]() {
								goto l2
							}
							depth--
							add(rulePegText, position7)
						}
						if !rules[ruleWSNLEOF]() {
							goto l2
						}
						{
							add(ruleAction0, position)
						}
						{
							position9 := position
							depth++
							{
								position10, tokenIndex10, depth10 := position, tokenIndex, depth
							l14:
								{
									position15, tokenIndex15, depth15 := position, tokenIndex, depth
									if !rules[ruleWS]() {
										goto l15
									}
									goto l14
								l15:
									position, tokenIndex, depth = position15, tokenIndex15, depth15
								}
								{
									position16, tokenIndex16, depth16 := position, tokenIndex, depth
									{
										position17, tokenIndex17, depth17 := position, tokenIndex, depth
										if buffer[position] != rune('B') {
											goto l18
										}
										position++
										if buffer[position] != rune('a') {
											goto l18
										}
										position++
										if buffer[position] != rune('c') {
											goto l18
										}
										position++
										if buffer[position] != rune('k') {
											goto l18
										}
										position++
										if buffer[position] != rune('g') {
											goto l18
										}
										position++
										if buffer[position] != rune('r') {
											goto l18
										}
										position++
										if buffer[position] != rune('o') {
											goto l18
										}
										position++
										if buffer[position] != rune('u') {
											goto l18
										}
										position++
										if buffer[position] != rune('n') {
											goto l18
										}
										position++
										if buffer[position] != rune('d') {
											goto l18
										}
										position++
										if buffer[position] != rune(':') {
											goto l18
										}
										position++
										goto l17
									l18:
										position, tokenIndex, depth = position17, tokenIndex17, depth17
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
										goto l17
									l19:
										position, tokenIndex, depth = position17, tokenIndex17, depth17
										if buffer[position] != rune('S') {
											goto l16
										}
										position++
										if buffer[position] != rune('c') {
											goto l16
										}
										position++
										if buffer[position] != rune('e') {
											goto l16
										}
										position++
										if buffer[position] != rune('n') {
											goto l16
										}
										position++
										if buffer[position] != rune('a') {
											goto l16
										}
										position++
										if buffer[position] != rune('r') {
											goto l16
										}
										position++
										if buffer[position] != rune('i') {
											goto l16
										}
										position++
										if buffer[position] != rune('o') {
											goto l16
										}
										position++
										if buffer[position] != rune(' ') {
											goto l16
										}
										position++
										if buffer[position] != rune('O') {
											goto l16
										}
										position++
										if buffer[position] != rune('u') {
											goto l16
										}
										position++
										if buffer[position] != rune('t') {
											goto l16
										}
										position++
										if buffer[position] != rune('l') {
											goto l16
										}
										position++
										if buffer[position] != rune('i') {
											goto l16
										}
										position++
										if buffer[position] != rune('n') {
											goto l16
										}
										position++
										if buffer[position] != rune('e') {
											goto l16
										}
										position++
										if buffer[position] != rune(':') {
											goto l16
										}
										position++
									}
								l17:
									goto l10
								l16:
									position, tokenIndex, depth = position16, tokenIndex16, depth16
								}
								if !rules[ruleStringToEndOfLine]() {
									goto l10
								}
								if !rules[ruleWSNLEOF]() {
									goto l10
								}
							l12:
								{
									position13, tokenIndex13, depth13 := position, tokenIndex, depth
								l20:
									{
										position21, tokenIndex21, depth21 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l21
										}
										goto l20
									l21:
										position, tokenIndex, depth = position21, tokenIndex21, depth21
									}
									{
										position22, tokenIndex22, depth22 := position, tokenIndex, depth
										{
											position23, tokenIndex23, depth23 := position, tokenIndex, depth
											if buffer[position] != rune('B') {
												goto l24
											}
											position++
											if buffer[position] != rune('a') {
												goto l24
											}
											position++
											if buffer[position] != rune('c') {
												goto l24
											}
											position++
											if buffer[position] != rune('k') {
												goto l24
											}
											position++
											if buffer[position] != rune('g') {
												goto l24
											}
											position++
											if buffer[position] != rune('r') {
												goto l24
											}
											position++
											if buffer[position] != rune('o') {
												goto l24
											}
											position++
											if buffer[position] != rune('u') {
												goto l24
											}
											position++
											if buffer[position] != rune('n') {
												goto l24
											}
											position++
											if buffer[position] != rune('d') {
												goto l24
											}
											position++
											if buffer[position] != rune(':') {
												goto l24
											}
											position++
											goto l23
										l24:
											position, tokenIndex, depth = position23, tokenIndex23, depth23
											if buffer[position] != rune('S') {
												goto l25
											}
											position++
											if buffer[position] != rune('c') {
												goto l25
											}
											position++
											if buffer[position] != rune('e') {
												goto l25
											}
											position++
											if buffer[position] != rune('n') {
												goto l25
											}
											position++
											if buffer[position] != rune('a') {
												goto l25
											}
											position++
											if buffer[position] != rune('r') {
												goto l25
											}
											position++
											if buffer[position] != rune('i') {
												goto l25
											}
											position++
											if buffer[position] != rune('o') {
												goto l25
											}
											position++
											if buffer[position] != rune(':') {
												goto l25
											}
											position++
											goto l23
										l25:
											position, tokenIndex, depth = position23, tokenIndex23, depth23
											if buffer[position] != rune('S') {
												goto l22
											}
											position++
											if buffer[position] != rune('c') {
												goto l22
											}
											position++
											if buffer[position] != rune('e') {
												goto l22
											}
											position++
											if buffer[position] != rune('n') {
												goto l22
											}
											position++
											if buffer[position] != rune('a') {
												goto l22
											}
											position++
											if buffer[position] != rune('r') {
												goto l22
											}
											position++
											if buffer[position] != rune('i') {
												goto l22
											}
											position++
											if buffer[position] != rune('o') {
												goto l22
											}
											position++
											if buffer[position] != rune(' ') {
												goto l22
											}
											position++
											if buffer[position] != rune('O') {
												goto l22
											}
											position++
											if buffer[position] != rune('u') {
												goto l22
											}
											position++
											if buffer[position] != rune('t') {
												goto l22
											}
											position++
											if buffer[position] != rune('l') {
												goto l22
											}
											position++
											if buffer[position] != rune('i') {
												goto l22
											}
											position++
											if buffer[position] != rune('n') {
												goto l22
											}
											position++
											if buffer[position] != rune('e') {
												goto l22
											}
											position++
											if buffer[position] != rune(':') {
												goto l22
											}
											position++
										}
									l23:
										goto l13
									l22:
										position, tokenIndex, depth = position22, tokenIndex22, depth22
									}
									if !rules[ruleStringToEndOfLine]() {
										goto l13
									}
									if !rules[ruleWSNLEOF]() {
										goto l13
									}
									goto l12
								l13:
									position, tokenIndex, depth = position13, tokenIndex13, depth13
								}
								goto l11
							l10:
								position, tokenIndex, depth = position10, tokenIndex10, depth10
							}
						l11:
							depth--
							add(rulePegText, position9)
						}
						{
							add(ruleAction1, position)
						}
					l27:
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							{
								position29, tokenIndex29, depth29 := position, tokenIndex, depth
								{
									position31 := position
									depth++
									if !rules[ruleTags]() {
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
										if !rules[ruleWS]() {
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
											if !rules[ruleStringToEndOfLine]() {
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
									if !rules[ruleWSNLEOF]() {
										goto l30
									}
									{
										add(ruleAction3, position)
									}
								l38:
									{
										position39, tokenIndex39, depth39 := position, tokenIndex, depth
										if !rules[ruleStep]() {
											goto l39
										}
										goto l38
									l39:
										position, tokenIndex, depth = position39, tokenIndex39, depth39
									}
									{
										add(ruleAction4, position)
									}
									depth--
									add(ruleBackground, position31)
								}
								goto l29
							l30:
								position, tokenIndex, depth = position29, tokenIndex29, depth29
								{
									position42 := position
									depth++
									if !rules[ruleTags]() {
										goto l41
									}
									if buffer[position] != rune('S') {
										goto l41
									}
									position++
									if buffer[position] != rune('c') {
										goto l41
									}
									position++
									if buffer[position] != rune('e') {
										goto l41
									}
									position++
									if buffer[position] != rune('n') {
										goto l41
									}
									position++
									if buffer[position] != rune('a') {
										goto l41
									}
									position++
									if buffer[position] != rune('r') {
										goto l41
									}
									position++
									if buffer[position] != rune('i') {
										goto l41
									}
									position++
									if buffer[position] != rune('o') {
										goto l41
									}
									position++
									if buffer[position] != rune(':') {
										goto l41
									}
									position++
								l43:
									{
										position44, tokenIndex44, depth44 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l44
										}
										goto l43
									l44:
										position, tokenIndex, depth = position44, tokenIndex44, depth44
									}
									{
										position45 := position
										depth++
										{
											position46, tokenIndex46, depth46 := position, tokenIndex, depth
											if !rules[ruleStringToEndOfLine]() {
												goto l46
											}
											goto l47
										l46:
											position, tokenIndex, depth = position46, tokenIndex46, depth46
										}
									l47:
										depth--
										add(rulePegText, position45)
									}
									if !rules[ruleWSNLEOF]() {
										goto l41
									}
									{
										add(ruleAction5, position)
									}
								l49:
									{
										position50, tokenIndex50, depth50 := position, tokenIndex, depth
										if !rules[ruleStep]() {
											goto l50
										}
										goto l49
									l50:
										position, tokenIndex, depth = position50, tokenIndex50, depth50
									}
									{
										add(ruleAction6, position)
									}
									depth--
									add(ruleScenario, position42)
								}
								goto l29
							l41:
								position, tokenIndex, depth = position29, tokenIndex29, depth29
								{
									position53 := position
									depth++
									if !rules[ruleTags]() {
										goto l52
									}
									if buffer[position] != rune('S') {
										goto l52
									}
									position++
									if buffer[position] != rune('c') {
										goto l52
									}
									position++
									if buffer[position] != rune('e') {
										goto l52
									}
									position++
									if buffer[position] != rune('n') {
										goto l52
									}
									position++
									if buffer[position] != rune('a') {
										goto l52
									}
									position++
									if buffer[position] != rune('r') {
										goto l52
									}
									position++
									if buffer[position] != rune('i') {
										goto l52
									}
									position++
									if buffer[position] != rune('o') {
										goto l52
									}
									position++
									if buffer[position] != rune(' ') {
										goto l52
									}
									position++
									if buffer[position] != rune('O') {
										goto l52
									}
									position++
									if buffer[position] != rune('u') {
										goto l52
									}
									position++
									if buffer[position] != rune('t') {
										goto l52
									}
									position++
									if buffer[position] != rune('l') {
										goto l52
									}
									position++
									if buffer[position] != rune('i') {
										goto l52
									}
									position++
									if buffer[position] != rune('n') {
										goto l52
									}
									position++
									if buffer[position] != rune('e') {
										goto l52
									}
									position++
									if buffer[position] != rune(':') {
										goto l52
									}
									position++
								l54:
									{
										position55, tokenIndex55, depth55 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l55
										}
										goto l54
									l55:
										position, tokenIndex, depth = position55, tokenIndex55, depth55
									}
									{
										position56 := position
										depth++
										{
											position57, tokenIndex57, depth57 := position, tokenIndex, depth
											if !rules[ruleStringToEndOfLine]() {
												goto l57
											}
											goto l58
										l57:
											position, tokenIndex, depth = position57, tokenIndex57, depth57
										}
									l58:
										depth--
										add(rulePegText, position56)
									}
									if !rules[ruleWSNLEOF]() {
										goto l52
									}
									{
										add(ruleAction7, position)
									}
								l60:
									{
										position61, tokenIndex61, depth61 := position, tokenIndex, depth
										if !rules[ruleStep]() {
											goto l61
										}
										goto l60
									l61:
										position, tokenIndex, depth = position61, tokenIndex61, depth61
									}
									{
										position62, tokenIndex62, depth62 := position, tokenIndex, depth
										{
											position64 := position
											depth++
											if !rules[ruleOS]() {
												goto l62
											}
											if buffer[position] != rune('E') {
												goto l62
											}
											position++
											if buffer[position] != rune('x') {
												goto l62
											}
											position++
											if buffer[position] != rune('a') {
												goto l62
											}
											position++
											if buffer[position] != rune('m') {
												goto l62
											}
											position++
											if buffer[position] != rune('p') {
												goto l62
											}
											position++
											if buffer[position] != rune('l') {
												goto l62
											}
											position++
											if buffer[position] != rune('e') {
												goto l62
											}
											position++
											if buffer[position] != rune('s') {
												goto l62
											}
											position++
											if buffer[position] != rune(':') {
												goto l62
											}
											position++
											if !rules[ruleWSNLEOF]() {
												goto l62
											}
											{
												add(ruleAction9, position)
											}
											{
												position66, tokenIndex66, depth66 := position, tokenIndex, depth
												if !rules[ruleTable]() {
													goto l66
												}
												goto l67
											l66:
												position, tokenIndex, depth = position66, tokenIndex66, depth66
											}
										l67:
											{
												add(ruleAction10, position)
											}
											depth--
											add(ruleOutlineExamples, position64)
										}
										goto l63
									l62:
										position, tokenIndex, depth = position62, tokenIndex62, depth62
									}
								l63:
									{
										add(ruleAction8, position)
									}
									depth--
									add(ruleOutline, position53)
								}
								goto l29
							l52:
								position, tokenIndex, depth = position29, tokenIndex29, depth29
								{
									position70 := position
									depth++
								l73:
									{
										position74, tokenIndex74, depth74 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l74
										}
										goto l73
									l74:
										position, tokenIndex, depth = position74, tokenIndex74, depth74
									}
									if !rules[ruleNL]() {
										goto l28
									}
								l71:
									{
										position72, tokenIndex72, depth72 := position, tokenIndex, depth
									l75:
										{
											position76, tokenIndex76, depth76 := position, tokenIndex, depth
											if !rules[ruleWS]() {
												goto l76
											}
											goto l75
										l76:
											position, tokenIndex, depth = position76, tokenIndex76, depth76
										}
										if !rules[ruleNL]() {
											goto l72
										}
										goto l71
									l72:
										position, tokenIndex, depth = position72, tokenIndex72, depth72
									}
									depth--
									add(ruleMLWS, position70)
								}
							}
						l29:
							goto l27
						l28:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
						}
						{
							add(ruleAction2, position)
						}
						depth--
						add(ruleFeature, position4)
					}
					goto l3
				l2:
					position, tokenIndex, depth = position2, tokenIndex2, depth2
				}
			l3:
				if !rules[ruleOS]() {
					goto l0
				}
				{
					position78, tokenIndex78, depth78 := position, tokenIndex, depth
					if !matchDot() {
						goto l78
					}
					goto l0
				l78:
					position, tokenIndex, depth = position78, tokenIndex78, depth78
				}
				depth--
				add(ruleBegin, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Feature <- <(Tags ('F' 'e' 'a' 't' 'u' 'r' 'e' ':') WS* <StringToEndOfLine> WSNLEOF Action0 <(WS* !(('B' 'a' 'c' 'k' 'g' 'r' 'o' 'u' 'n' 'd' ':') / ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ':') / ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':')) StringToEndOfLine WSNLEOF)+?> Action1 (Background / Scenario / Outline / MLWS)* Action2)> */
		nil,
		/* 2 Background <- <(Tags ('B' 'a' 'c' 'k' 'g' 'r' 'o' 'u' 'n' 'd' ':') WS* <StringToEndOfLine?> WSNLEOF Action3 Step* Action4)> */
		nil,
		/* 3 Scenario <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ':') WS* <StringToEndOfLine?> WSNLEOF Action5 Step* Action6)> */
		nil,
		/* 4 Outline <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':') WS* <StringToEndOfLine?> WSNLEOF Action7 Step* OutlineExamples? Action8)> */
		nil,
		/* 5 OutlineExamples <- <(OS ('E' 'x' 'a' 'm' 'p' 'l' 'e' 's' ':') WSNLEOF Action9 Table? Action10)> */
		nil,
		/* 6 Step <- <(Tags <((&('B') ('B' 'u' 't')) | (&('O') ('O' 'r')) | (&('A') ('A' 'n' 'd')) | (&('T') ('T' 'h' 'e' 'n')) | (&('W') ('W' 'h' 'e' 'n')) | (&('G') ('G' 'i' 'v' 'e' 'n')))> Action11 WS* <StringToEndOfLine> WSNLEOF Action12 StepArgument? Action13)> */
		func() bool {
			position84, tokenIndex84, depth84 := position, tokenIndex, depth
			{
				position85 := position
				depth++
				if !rules[ruleTags]() {
					goto l84
				}
				{
					position86 := position
					depth++
					{
						switch buffer[position] {
						case 'B':
							if buffer[position] != rune('B') {
								goto l84
							}
							position++
							if buffer[position] != rune('u') {
								goto l84
							}
							position++
							if buffer[position] != rune('t') {
								goto l84
							}
							position++
							break
						case 'O':
							if buffer[position] != rune('O') {
								goto l84
							}
							position++
							if buffer[position] != rune('r') {
								goto l84
							}
							position++
							break
						case 'A':
							if buffer[position] != rune('A') {
								goto l84
							}
							position++
							if buffer[position] != rune('n') {
								goto l84
							}
							position++
							if buffer[position] != rune('d') {
								goto l84
							}
							position++
							break
						case 'T':
							if buffer[position] != rune('T') {
								goto l84
							}
							position++
							if buffer[position] != rune('h') {
								goto l84
							}
							position++
							if buffer[position] != rune('e') {
								goto l84
							}
							position++
							if buffer[position] != rune('n') {
								goto l84
							}
							position++
							break
						case 'W':
							if buffer[position] != rune('W') {
								goto l84
							}
							position++
							if buffer[position] != rune('h') {
								goto l84
							}
							position++
							if buffer[position] != rune('e') {
								goto l84
							}
							position++
							if buffer[position] != rune('n') {
								goto l84
							}
							position++
							break
						default:
							if buffer[position] != rune('G') {
								goto l84
							}
							position++
							if buffer[position] != rune('i') {
								goto l84
							}
							position++
							if buffer[position] != rune('v') {
								goto l84
							}
							position++
							if buffer[position] != rune('e') {
								goto l84
							}
							position++
							if buffer[position] != rune('n') {
								goto l84
							}
							position++
							break
						}
					}

					depth--
					add(rulePegText, position86)
				}
				{
					add(ruleAction11, position)
				}
			l89:
				{
					position90, tokenIndex90, depth90 := position, tokenIndex, depth
					if !rules[ruleWS]() {
						goto l90
					}
					goto l89
				l90:
					position, tokenIndex, depth = position90, tokenIndex90, depth90
				}
				{
					position91 := position
					depth++
					if !rules[ruleStringToEndOfLine]() {
						goto l84
					}
					depth--
					add(rulePegText, position91)
				}
				if !rules[ruleWSNLEOF]() {
					goto l84
				}
				{
					add(ruleAction12, position)
				}
				{
					position93, tokenIndex93, depth93 := position, tokenIndex, depth
					{
						position95 := position
						depth++
						{
							position96, tokenIndex96, depth96 := position, tokenIndex, depth
							if !rules[ruleTable]() {
								goto l97
							}
							goto l96
						l97:
							position, tokenIndex, depth = position96, tokenIndex96, depth96
							{
								position98 := position
								depth++
							l99:
								{
									position100, tokenIndex100, depth100 := position, tokenIndex, depth
								l101:
									{
										position102, tokenIndex102, depth102 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l102
										}
										goto l101
									l102:
										position, tokenIndex, depth = position102, tokenIndex102, depth102
									}
									if !rules[ruleNL]() {
										goto l100
									}
									goto l99
								l100:
									position, tokenIndex, depth = position100, tokenIndex100, depth100
								}
								{
									position103 := position
									depth++
								l104:
									{
										position105, tokenIndex105, depth105 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l105
										}
										goto l104
									l105:
										position, tokenIndex, depth = position105, tokenIndex105, depth105
									}
									depth--
									add(rulePegText, position103)
								}
								if !rules[rulePyStringQuote]() {
									goto l93
								}
								if !rules[ruleNL]() {
									goto l93
								}
								{
									add(ruleAction14, position)
								}
							l107:
								{
									position108, tokenIndex108, depth108 := position, tokenIndex, depth
									{
										position109, tokenIndex109, depth109 := position, tokenIndex, depth
									l110:
										{
											position111, tokenIndex111, depth111 := position, tokenIndex, depth
											if !rules[ruleWS]() {
												goto l111
											}
											goto l110
										l111:
											position, tokenIndex, depth = position111, tokenIndex111, depth111
										}
										if !rules[rulePyStringQuote]() {
											goto l109
										}
										goto l108
									l109:
										position, tokenIndex, depth = position109, tokenIndex109, depth109
									}
									{
										position112 := position
										depth++
										{
											position113 := position
											depth++
											if !rules[ruleStringToEndOfLine]() {
												goto l108
											}
											if !rules[ruleNL]() {
												goto l108
											}
											depth--
											add(rulePegText, position113)
										}
										{
											add(ruleAction16, position)
										}
										depth--
										add(rulePyStringLine, position112)
									}
									goto l107
								l108:
									position, tokenIndex, depth = position108, tokenIndex108, depth108
								}
							l115:
								{
									position116, tokenIndex116, depth116 := position, tokenIndex, depth
									if !rules[ruleWS]() {
										goto l116
									}
									goto l115
								l116:
									position, tokenIndex, depth = position116, tokenIndex116, depth116
								}
								if !rules[rulePyStringQuote]() {
									goto l93
								}
								if !rules[ruleWSNLEOF]() {
									goto l93
								}
								{
									add(ruleAction15, position)
								}
								depth--
								add(rulePyString, position98)
							}
						}
					l96:
						depth--
						add(ruleStepArgument, position95)
					}
					goto l94
				l93:
					position, tokenIndex, depth = position93, tokenIndex93, depth93
				}
			l94:
				{
					add(ruleAction13, position)
				}
				depth--
				add(ruleStep, position85)
			}
			return true
		l84:
			position, tokenIndex, depth = position84, tokenIndex84, depth84
			return false
		},
		/* 7 StepArgument <- <(Table / PyString)> */
		nil,
		/* 8 PyString <- <((WS* NL)* <WS*> PyStringQuote NL Action14 (!(WS* PyStringQuote) PyStringLine)* WS* PyStringQuote WSNLEOF Action15)> */
		nil,
		/* 9 PyStringQuote <- <('"' '"' '"')> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				if buffer[position] != rune('"') {
					goto l121
				}
				position++
				if buffer[position] != rune('"') {
					goto l121
				}
				position++
				if buffer[position] != rune('"') {
					goto l121
				}
				position++
				depth--
				add(rulePyStringQuote, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 10 PyStringLine <- <(<(StringToEndOfLine NL)> Action16)> */
		nil,
		/* 11 Table <- <(Action17 TableRow+ Action18)> */
		func() bool {
			position124, tokenIndex124, depth124 := position, tokenIndex, depth
			{
				position125 := position
				depth++
				{
					add(ruleAction17, position)
				}
				{
					position129 := position
					depth++
					{
						add(ruleAction19, position)
					}
					if !rules[ruleOS]() {
						goto l124
					}
					if buffer[position] != rune('|') {
						goto l124
					}
					position++
					{
						position133 := position
						depth++
						{
							position134 := position
							depth++
							{
								position137, tokenIndex137, depth137 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '|':
										if buffer[position] != rune('|') {
											goto l137
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l137
										}
										position++
										break
									default:
										if buffer[position] != rune('\r') {
											goto l137
										}
										position++
										break
									}
								}

								goto l124
							l137:
								position, tokenIndex, depth = position137, tokenIndex137, depth137
							}
							if !matchDot() {
								goto l124
							}
						l135:
							{
								position136, tokenIndex136, depth136 := position, tokenIndex, depth
								{
									position139, tokenIndex139, depth139 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l139
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l139
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l139
											}
											position++
											break
										}
									}

									goto l136
								l139:
									position, tokenIndex, depth = position139, tokenIndex139, depth139
								}
								if !matchDot() {
									goto l136
								}
								goto l135
							l136:
								position, tokenIndex, depth = position136, tokenIndex136, depth136
							}
							depth--
							add(rulePegText, position134)
						}
						if buffer[position] != rune('|') {
							goto l124
						}
						position++
						{
							add(ruleAction21, position)
						}
						depth--
						add(ruleTableCell, position133)
					}
				l131:
					{
						position132, tokenIndex132, depth132 := position, tokenIndex, depth
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

									goto l132
								l146:
									position, tokenIndex, depth = position146, tokenIndex146, depth146
								}
								if !matchDot() {
									goto l132
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
								goto l132
							}
							position++
							{
								add(ruleAction21, position)
							}
							depth--
							add(ruleTableCell, position142)
						}
						goto l131
					l132:
						position, tokenIndex, depth = position132, tokenIndex132, depth132
					}
					if !rules[ruleWSNLEOF]() {
						goto l124
					}
					{
						add(ruleAction20, position)
					}
					depth--
					add(ruleTableRow, position129)
				}
			l127:
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					{
						position152 := position
						depth++
						{
							add(ruleAction19, position)
						}
						if !rules[ruleOS]() {
							goto l128
						}
						if buffer[position] != rune('|') {
							goto l128
						}
						position++
						{
							position156 := position
							depth++
							{
								position157 := position
								depth++
								{
									position160, tokenIndex160, depth160 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l160
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l160
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l160
											}
											position++
											break
										}
									}

									goto l128
								l160:
									position, tokenIndex, depth = position160, tokenIndex160, depth160
								}
								if !matchDot() {
									goto l128
								}
							l158:
								{
									position159, tokenIndex159, depth159 := position, tokenIndex, depth
									{
										position162, tokenIndex162, depth162 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l162
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l162
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l162
												}
												position++
												break
											}
										}

										goto l159
									l162:
										position, tokenIndex, depth = position162, tokenIndex162, depth162
									}
									if !matchDot() {
										goto l159
									}
									goto l158
								l159:
									position, tokenIndex, depth = position159, tokenIndex159, depth159
								}
								depth--
								add(rulePegText, position157)
							}
							if buffer[position] != rune('|') {
								goto l128
							}
							position++
							{
								add(ruleAction21, position)
							}
							depth--
							add(ruleTableCell, position156)
						}
					l154:
						{
							position155, tokenIndex155, depth155 := position, tokenIndex, depth
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

										goto l155
									l169:
										position, tokenIndex, depth = position169, tokenIndex169, depth169
									}
									if !matchDot() {
										goto l155
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
									goto l155
								}
								position++
								{
									add(ruleAction21, position)
								}
								depth--
								add(ruleTableCell, position165)
							}
							goto l154
						l155:
							position, tokenIndex, depth = position155, tokenIndex155, depth155
						}
						if !rules[ruleWSNLEOF]() {
							goto l128
						}
						{
							add(ruleAction20, position)
						}
						depth--
						add(ruleTableRow, position152)
					}
					goto l127
				l128:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
				}
				{
					add(ruleAction18, position)
				}
				depth--
				add(ruleTable, position125)
			}
			return true
		l124:
			position, tokenIndex, depth = position124, tokenIndex124, depth124
			return false
		},
		/* 12 TableRow <- <(Action19 OS '|' TableCell+ WSNLEOF Action20)> */
		nil,
		/* 13 TableCell <- <(<(!((&('|') '|') | (&('\n') '\n') | (&('\r') '\r')) .)+> '|' Action21)> */
		nil,
		/* 14 Tags <- <(Tag* OS)> */
		func() bool {
			position178, tokenIndex178, depth178 := position, tokenIndex, depth
			{
				position179 := position
				depth++
			l180:
				{
					position181, tokenIndex181, depth181 := position, tokenIndex, depth
					{
						position182 := position
						depth++
						if !rules[ruleOS]() {
							goto l181
						}
						if buffer[position] != rune('@') {
							goto l181
						}
						position++
						{
							position183 := position
							depth++
							{
								position186, tokenIndex186, depth186 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case ' ':
										if buffer[position] != rune(' ') {
											goto l186
										}
										position++
										break
									case '\t':
										if buffer[position] != rune('\t') {
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

								goto l181
							l186:
								position, tokenIndex, depth = position186, tokenIndex186, depth186
							}
							if !matchDot() {
								goto l181
							}
						l184:
							{
								position185, tokenIndex185, depth185 := position, tokenIndex, depth
								{
									position188, tokenIndex188, depth188 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case ' ':
											if buffer[position] != rune(' ') {
												goto l188
											}
											position++
											break
										case '\t':
											if buffer[position] != rune('\t') {
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
						{
							add(ruleAction22, position)
						}
						depth--
						add(ruleTag, position182)
					}
					goto l180
				l181:
					position, tokenIndex, depth = position181, tokenIndex181, depth181
				}
				if !rules[ruleOS]() {
					goto l178
				}
				depth--
				add(ruleTags, position179)
			}
			return true
		l178:
			position, tokenIndex, depth = position178, tokenIndex178, depth178
			return false
		},
		/* 15 Tag <- <(OS '@' <(!((&(' ') ' ') | (&('\t') '\t') | (&('\n') '\n') | (&('\r') '\r')) .)+> Action22)> */
		nil,
		/* 16 StringToEndOfLine <- <(!'\n' .)+> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				{
					position196, tokenIndex196, depth196 := position, tokenIndex, depth
					if buffer[position] != rune('\n') {
						goto l196
					}
					position++
					goto l192
				l196:
					position, tokenIndex, depth = position196, tokenIndex196, depth196
				}
				if !matchDot() {
					goto l192
				}
			l194:
				{
					position195, tokenIndex195, depth195 := position, tokenIndex, depth
					{
						position197, tokenIndex197, depth197 := position, tokenIndex, depth
						if buffer[position] != rune('\n') {
							goto l197
						}
						position++
						goto l195
					l197:
						position, tokenIndex, depth = position197, tokenIndex197, depth197
					}
					if !matchDot() {
						goto l195
					}
					goto l194
				l195:
					position, tokenIndex, depth = position195, tokenIndex195, depth195
				}
				depth--
				add(ruleStringToEndOfLine, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 17 OS <- <(NL / WS)*> */
		func() bool {
			{
				position199 := position
				depth++
			l200:
				{
					position201, tokenIndex201, depth201 := position, tokenIndex, depth
					{
						position202, tokenIndex202, depth202 := position, tokenIndex, depth
						if !rules[ruleNL]() {
							goto l203
						}
						goto l202
					l203:
						position, tokenIndex, depth = position202, tokenIndex202, depth202
						if !rules[ruleWS]() {
							goto l201
						}
					}
				l202:
					goto l200
				l201:
					position, tokenIndex, depth = position201, tokenIndex201, depth201
				}
				depth--
				add(ruleOS, position199)
			}
			return true
		},
		/* 18 WS <- <(' ' / '\t')> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				{
					position206, tokenIndex206, depth206 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l207
					}
					position++
					goto l206
				l207:
					position, tokenIndex, depth = position206, tokenIndex206, depth206
					if buffer[position] != rune('\t') {
						goto l204
					}
					position++
				}
			l206:
				depth--
				add(ruleWS, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 19 NL <- <('\n' / '\r' / ('\r' '\n'))> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				{
					position210, tokenIndex210, depth210 := position, tokenIndex, depth
					if buffer[position] != rune('\n') {
						goto l211
					}
					position++
					goto l210
				l211:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
					if buffer[position] != rune('\r') {
						goto l212
					}
					position++
					goto l210
				l212:
					position, tokenIndex, depth = position210, tokenIndex210, depth210
					if buffer[position] != rune('\r') {
						goto l208
					}
					position++
					if buffer[position] != rune('\n') {
						goto l208
					}
					position++
				}
			l210:
				depth--
				add(ruleNL, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 20 WSNLEOF <- <(WS* (NL / !.))> */
		func() bool {
			position213, tokenIndex213, depth213 := position, tokenIndex, depth
			{
				position214 := position
				depth++
			l215:
				{
					position216, tokenIndex216, depth216 := position, tokenIndex, depth
					if !rules[ruleWS]() {
						goto l216
					}
					goto l215
				l216:
					position, tokenIndex, depth = position216, tokenIndex216, depth216
				}
				{
					position217, tokenIndex217, depth217 := position, tokenIndex, depth
					if !rules[ruleNL]() {
						goto l218
					}
					goto l217
				l218:
					position, tokenIndex, depth = position217, tokenIndex217, depth217
					{
						position219, tokenIndex219, depth219 := position, tokenIndex, depth
						if !matchDot() {
							goto l219
						}
						goto l213
					l219:
						position, tokenIndex, depth = position219, tokenIndex219, depth219
					}
				}
			l217:
				depth--
				add(ruleWSNLEOF, position214)
			}
			return true
		l213:
			position, tokenIndex, depth = position213, tokenIndex213, depth213
			return false
		},
		/* 21 MLWS <- <(WS* NL)+> */
		nil,
		nil,
		/* 24 Action0 <- <{ p.buffer2 = buffer[begin:end] }> */
		nil,
		/* 25 Action1 <- <{ p.beginFeature(p.buffer2, buffer[begin:end], p.tagBuffer); p.tagBuffer = nil }> */
		nil,
		/* 26 Action2 <- <{ p.endFeature() }> */
		nil,
		/* 27 Action3 <- <{ p.beginBackground(buffer[begin:end], p.tagBuffer); p.tagBuffer = nil }> */
		nil,
		/* 28 Action4 <- <{ p.endBackground() }> */
		nil,
		/* 29 Action5 <- <{ p.beginScenario(buffer[begin:end], p.tagBuffer); p.tagBuffer = nil }> */
		nil,
		/* 30 Action6 <- <{ p.endScenario() }> */
		nil,
		/* 31 Action7 <- <{ p.beginOutline(buffer[begin:end], p.tagBuffer); p.tagBuffer = nil }> */
		nil,
		/* 32 Action8 <- <{ p.endOutline() }> */
		nil,
		/* 33 Action9 <- <{ p.beginOutlineExamples() }> */
		nil,
		/* 34 Action10 <- <{ p.endOutlineExamples() }> */
		nil,
		/* 35 Action11 <- <{ p.buffer2 = buffer[begin:end] }> */
		nil,
		/* 36 Action12 <- <{ p.beginStep(p.buffer2, buffer[begin:end]) }> */
		nil,
		/* 37 Action13 <- <{ p.endStep() }> */
		nil,
		/* 38 Action14 <- <{ p.beginPyString(buffer[begin:end]) }> */
		nil,
		/* 39 Action15 <- <{ p.endPyString() }> */
		nil,
		/* 40 Action16 <- <{ p.bufferPyString(buffer[begin:end]) }> */
		nil,
		/* 41 Action17 <- <{ p.beginTable() }> */
		nil,
		/* 42 Action18 <- <{ p.endTable() }> */
		nil,
		/* 43 Action19 <- <{ p.beginTableRow() }> */
		nil,
		/* 44 Action20 <- <{ p.endTableRow() }> */
		nil,
		/* 45 Action21 <- <{ p.beginTableCell(); p.endTableCell(buffer[begin:end]) }> */
		nil,
		/* 46 Action22 <- <{ p.tagBuffer = append(p.tagBuffer, buffer[begin:end]) }> */
		nil,
	}
	p.rules = rules
}
