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

	buf1    string
	buf2    string
	buf3    string
	buftags []string
	bufcmt  string

	Buffer string
	buffer []rune
	rules  [60]func() bool
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
			p.beginBackground(p.buf1, p.buftags)
			p.buftags = nil
		case ruleAction8:
			p.endBackground()
		case ruleAction9:
			p.buf1 = buffer[begin:end]
		case ruleAction10:
			p.beginScenario(p.buf1, p.buftags)
			p.buftags = nil
		case ruleAction11:
			p.endScenario()
		case ruleAction12:
			p.buf1 = buffer[begin:end]
		case ruleAction13:
			p.beginOutline(p.buf1, p.buftags)
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
			p.beginStep(p.buf1, p.buf2)
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
			p.endTableCell(buffer[begin:end])
		case ruleAction29:
			p.buftags = append(p.buftags, buffer[begin:end])
		case ruleAction30:
			p.bufcmt = buffer[begin:end]

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
							{
								position8, tokenIndex8, depth8 := position, tokenIndex, depth
								if !rules[ruleUntilLineEnd]() {
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
						if !rules[ruleLineEnd]() {
							goto l2
						}
						{
							position13, tokenIndex13, depth13 := position, tokenIndex, depth
						l17:
							{
								position18, tokenIndex18, depth18 := position, tokenIndex, depth
								if !rules[ruleWS]() {
									goto l18
								}
								goto l17
							l18:
								position, tokenIndex, depth = position18, tokenIndex18, depth18
							}
							{
								position19, tokenIndex19, depth19 := position, tokenIndex, depth
								{
									position20, tokenIndex20, depth20 := position, tokenIndex, depth
									if buffer[position] != rune('B') {
										goto l21
									}
									position++
									if buffer[position] != rune('a') {
										goto l21
									}
									position++
									if buffer[position] != rune('c') {
										goto l21
									}
									position++
									if buffer[position] != rune('k') {
										goto l21
									}
									position++
									if buffer[position] != rune('g') {
										goto l21
									}
									position++
									if buffer[position] != rune('r') {
										goto l21
									}
									position++
									if buffer[position] != rune('o') {
										goto l21
									}
									position++
									if buffer[position] != rune('u') {
										goto l21
									}
									position++
									if buffer[position] != rune('n') {
										goto l21
									}
									position++
									if buffer[position] != rune('d') {
										goto l21
									}
									position++
									if buffer[position] != rune(':') {
										goto l21
									}
									position++
									goto l20
								l21:
									position, tokenIndex, depth = position20, tokenIndex20, depth20
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
									if buffer[position] != rune(':') {
										goto l22
									}
									position++
									goto l20
								l22:
									position, tokenIndex, depth = position20, tokenIndex20, depth20
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
									if buffer[position] != rune(' ') {
										goto l19
									}
									position++
									if buffer[position] != rune('O') {
										goto l19
									}
									position++
									if buffer[position] != rune('u') {
										goto l19
									}
									position++
									if buffer[position] != rune('t') {
										goto l19
									}
									position++
									if buffer[position] != rune('l') {
										goto l19
									}
									position++
									if buffer[position] != rune('i') {
										goto l19
									}
									position++
									if buffer[position] != rune('n') {
										goto l19
									}
									position++
									if buffer[position] != rune('e') {
										goto l19
									}
									position++
									if buffer[position] != rune(':') {
										goto l19
									}
									position++
								}
							l20:
								goto l13
							l19:
								position, tokenIndex, depth = position19, tokenIndex19, depth19
							}
							{
								position23 := position
								depth++
								if !rules[ruleUntilLineEnd]() {
									goto l13
								}
								depth--
								add(rulePegText, position23)
							}
							{
								add(ruleAction2, position)
							}
							if !rules[ruleLineEnd]() {
								goto l13
							}
							{
								add(ruleAction3, position)
							}
						l15:
							{
								position16, tokenIndex16, depth16 := position, tokenIndex, depth
							l26:
								{
									position27, tokenIndex27, depth27 := position, tokenIndex, depth
									if !rules[ruleWS]() {
										goto l27
									}
									goto l26
								l27:
									position, tokenIndex, depth = position27, tokenIndex27, depth27
								}
								{
									position28, tokenIndex28, depth28 := position, tokenIndex, depth
									{
										position29, tokenIndex29, depth29 := position, tokenIndex, depth
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
										goto l29
									l30:
										position, tokenIndex, depth = position29, tokenIndex29, depth29
										if buffer[position] != rune('S') {
											goto l31
										}
										position++
										if buffer[position] != rune('c') {
											goto l31
										}
										position++
										if buffer[position] != rune('e') {
											goto l31
										}
										position++
										if buffer[position] != rune('n') {
											goto l31
										}
										position++
										if buffer[position] != rune('a') {
											goto l31
										}
										position++
										if buffer[position] != rune('r') {
											goto l31
										}
										position++
										if buffer[position] != rune('i') {
											goto l31
										}
										position++
										if buffer[position] != rune('o') {
											goto l31
										}
										position++
										if buffer[position] != rune(':') {
											goto l31
										}
										position++
										goto l29
									l31:
										position, tokenIndex, depth = position29, tokenIndex29, depth29
										if buffer[position] != rune('S') {
											goto l28
										}
										position++
										if buffer[position] != rune('c') {
											goto l28
										}
										position++
										if buffer[position] != rune('e') {
											goto l28
										}
										position++
										if buffer[position] != rune('n') {
											goto l28
										}
										position++
										if buffer[position] != rune('a') {
											goto l28
										}
										position++
										if buffer[position] != rune('r') {
											goto l28
										}
										position++
										if buffer[position] != rune('i') {
											goto l28
										}
										position++
										if buffer[position] != rune('o') {
											goto l28
										}
										position++
										if buffer[position] != rune(' ') {
											goto l28
										}
										position++
										if buffer[position] != rune('O') {
											goto l28
										}
										position++
										if buffer[position] != rune('u') {
											goto l28
										}
										position++
										if buffer[position] != rune('t') {
											goto l28
										}
										position++
										if buffer[position] != rune('l') {
											goto l28
										}
										position++
										if buffer[position] != rune('i') {
											goto l28
										}
										position++
										if buffer[position] != rune('n') {
											goto l28
										}
										position++
										if buffer[position] != rune('e') {
											goto l28
										}
										position++
										if buffer[position] != rune(':') {
											goto l28
										}
										position++
									}
								l29:
									goto l16
								l28:
									position, tokenIndex, depth = position28, tokenIndex28, depth28
								}
								{
									position32 := position
									depth++
									if !rules[ruleUntilLineEnd]() {
										goto l16
									}
									depth--
									add(rulePegText, position32)
								}
								{
									add(ruleAction2, position)
								}
								if !rules[ruleLineEnd]() {
									goto l16
								}
								{
									add(ruleAction3, position)
								}
								goto l15
							l16:
								position, tokenIndex, depth = position16, tokenIndex16, depth16
							}
							goto l14
						l13:
							position, tokenIndex, depth = position13, tokenIndex13, depth13
						}
					l14:
						{
							add(ruleAction4, position)
						}
					l36:
						{
							position37, tokenIndex37, depth37 := position, tokenIndex, depth
							{
								position38, tokenIndex38, depth38 := position, tokenIndex, depth
								{
									position40 := position
									depth++
									if !rules[ruleTags]() {
										goto l39
									}
									if buffer[position] != rune('B') {
										goto l39
									}
									position++
									if buffer[position] != rune('a') {
										goto l39
									}
									position++
									if buffer[position] != rune('c') {
										goto l39
									}
									position++
									if buffer[position] != rune('k') {
										goto l39
									}
									position++
									if buffer[position] != rune('g') {
										goto l39
									}
									position++
									if buffer[position] != rune('r') {
										goto l39
									}
									position++
									if buffer[position] != rune('o') {
										goto l39
									}
									position++
									if buffer[position] != rune('u') {
										goto l39
									}
									position++
									if buffer[position] != rune('n') {
										goto l39
									}
									position++
									if buffer[position] != rune('d') {
										goto l39
									}
									position++
									if buffer[position] != rune(':') {
										goto l39
									}
									position++
								l41:
									{
										position42, tokenIndex42, depth42 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l42
										}
										goto l41
									l42:
										position, tokenIndex, depth = position42, tokenIndex42, depth42
									}
									{
										position43 := position
										depth++
										{
											position44, tokenIndex44, depth44 := position, tokenIndex, depth
											if !rules[ruleUntilLineEnd]() {
												goto l44
											}
											goto l45
										l44:
											position, tokenIndex, depth = position44, tokenIndex44, depth44
										}
									l45:
										depth--
										add(rulePegText, position43)
									}
									{
										add(ruleAction6, position)
									}
									if !rules[ruleLineEnd]() {
										goto l39
									}
									{
										add(ruleAction7, position)
									}
								l48:
									{
										position49, tokenIndex49, depth49 := position, tokenIndex, depth
										if !rules[ruleStep]() {
											goto l49
										}
										goto l48
									l49:
										position, tokenIndex, depth = position49, tokenIndex49, depth49
									}
									{
										add(ruleAction8, position)
									}
									depth--
									add(ruleBackground, position40)
								}
								goto l38
							l39:
								position, tokenIndex, depth = position38, tokenIndex38, depth38
								{
									position52 := position
									depth++
									if !rules[ruleTags]() {
										goto l51
									}
									if buffer[position] != rune('S') {
										goto l51
									}
									position++
									if buffer[position] != rune('c') {
										goto l51
									}
									position++
									if buffer[position] != rune('e') {
										goto l51
									}
									position++
									if buffer[position] != rune('n') {
										goto l51
									}
									position++
									if buffer[position] != rune('a') {
										goto l51
									}
									position++
									if buffer[position] != rune('r') {
										goto l51
									}
									position++
									if buffer[position] != rune('i') {
										goto l51
									}
									position++
									if buffer[position] != rune('o') {
										goto l51
									}
									position++
									if buffer[position] != rune(':') {
										goto l51
									}
									position++
								l53:
									{
										position54, tokenIndex54, depth54 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l54
										}
										goto l53
									l54:
										position, tokenIndex, depth = position54, tokenIndex54, depth54
									}
									{
										position55 := position
										depth++
										{
											position56, tokenIndex56, depth56 := position, tokenIndex, depth
											if !rules[ruleUntilLineEnd]() {
												goto l56
											}
											goto l57
										l56:
											position, tokenIndex, depth = position56, tokenIndex56, depth56
										}
									l57:
										depth--
										add(rulePegText, position55)
									}
									{
										add(ruleAction9, position)
									}
									if !rules[ruleLineEnd]() {
										goto l51
									}
									{
										add(ruleAction10, position)
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
										add(ruleAction11, position)
									}
									depth--
									add(ruleScenario, position52)
								}
								goto l38
							l51:
								position, tokenIndex, depth = position38, tokenIndex38, depth38
								{
									position64 := position
									depth++
									if !rules[ruleTags]() {
										goto l63
									}
									if buffer[position] != rune('S') {
										goto l63
									}
									position++
									if buffer[position] != rune('c') {
										goto l63
									}
									position++
									if buffer[position] != rune('e') {
										goto l63
									}
									position++
									if buffer[position] != rune('n') {
										goto l63
									}
									position++
									if buffer[position] != rune('a') {
										goto l63
									}
									position++
									if buffer[position] != rune('r') {
										goto l63
									}
									position++
									if buffer[position] != rune('i') {
										goto l63
									}
									position++
									if buffer[position] != rune('o') {
										goto l63
									}
									position++
									if buffer[position] != rune(' ') {
										goto l63
									}
									position++
									if buffer[position] != rune('O') {
										goto l63
									}
									position++
									if buffer[position] != rune('u') {
										goto l63
									}
									position++
									if buffer[position] != rune('t') {
										goto l63
									}
									position++
									if buffer[position] != rune('l') {
										goto l63
									}
									position++
									if buffer[position] != rune('i') {
										goto l63
									}
									position++
									if buffer[position] != rune('n') {
										goto l63
									}
									position++
									if buffer[position] != rune('e') {
										goto l63
									}
									position++
									if buffer[position] != rune(':') {
										goto l63
									}
									position++
								l65:
									{
										position66, tokenIndex66, depth66 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l66
										}
										goto l65
									l66:
										position, tokenIndex, depth = position66, tokenIndex66, depth66
									}
									{
										position67 := position
										depth++
										{
											position68, tokenIndex68, depth68 := position, tokenIndex, depth
											if !rules[ruleUntilLineEnd]() {
												goto l68
											}
											goto l69
										l68:
											position, tokenIndex, depth = position68, tokenIndex68, depth68
										}
									l69:
										depth--
										add(rulePegText, position67)
									}
									{
										add(ruleAction12, position)
									}
									if !rules[ruleLineEnd]() {
										goto l63
									}
									{
										add(ruleAction13, position)
									}
								l72:
									{
										position73, tokenIndex73, depth73 := position, tokenIndex, depth
										if !rules[ruleStep]() {
											goto l73
										}
										goto l72
									l73:
										position, tokenIndex, depth = position73, tokenIndex73, depth73
									}
									{
										position74, tokenIndex74, depth74 := position, tokenIndex, depth
										{
											position76 := position
											depth++
											if !rules[ruleOS]() {
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
											if !rules[ruleLineEnd]() {
												goto l74
											}
											{
												add(ruleAction15, position)
											}
											{
												position78, tokenIndex78, depth78 := position, tokenIndex, depth
												if !rules[ruleTable]() {
													goto l78
												}
												goto l79
											l78:
												position, tokenIndex, depth = position78, tokenIndex78, depth78
											}
										l79:
											{
												add(ruleAction16, position)
											}
											depth--
											add(ruleOutlineExamples, position76)
										}
										goto l75
									l74:
										position, tokenIndex, depth = position74, tokenIndex74, depth74
									}
								l75:
									{
										add(ruleAction14, position)
									}
									depth--
									add(ruleOutline, position64)
								}
								goto l38
							l63:
								position, tokenIndex, depth = position38, tokenIndex38, depth38
								{
									position82 := position
									depth++
									if !rules[ruleWS]() {
										goto l37
									}
									if !rules[ruleLineEnd]() {
										goto l37
									}
								l83:
									{
										position84, tokenIndex84, depth84 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l84
										}
										if !rules[ruleLineEnd]() {
											goto l84
										}
										goto l83
									l84:
										position, tokenIndex, depth = position84, tokenIndex84, depth84
									}
									depth--
									add(ruleBlankLine, position82)
								}
							}
						l38:
							goto l36
						l37:
							position, tokenIndex, depth = position37, tokenIndex37, depth37
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
				if !rules[ruleOS]() {
					goto l0
				}
				{
					position86, tokenIndex86, depth86 := position, tokenIndex, depth
					if !matchDot() {
						goto l86
					}
					goto l0
				l86:
					position, tokenIndex, depth = position86, tokenIndex86, depth86
				}
				depth--
				add(ruleBegin, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Feature <- <(Tags ('F' 'e' 'a' 't' 'u' 'r' 'e' ':') WS* <UntilLineEnd?> Action0 <> Action1 LineEnd (WS* !(('B' 'a' 'c' 'k' 'g' 'r' 'o' 'u' 'n' 'd' ':') / ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ':') / ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':')) <UntilLineEnd> Action2 LineEnd Action3)+? Action4 (Background / Scenario / Outline / BlankLine)* Action5)> */
		nil,
		/* 2 Background <- <(Tags ('B' 'a' 'c' 'k' 'g' 'r' 'o' 'u' 'n' 'd' ':') WS* <UntilLineEnd?> Action6 LineEnd Action7 Step* Action8)> */
		nil,
		/* 3 Scenario <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ':') WS* <UntilLineEnd?> Action9 LineEnd Action10 Step* Action11)> */
		nil,
		/* 4 Outline <- <(Tags ('S' 'c' 'e' 'n' 'a' 'r' 'i' 'o' ' ' 'O' 'u' 't' 'l' 'i' 'n' 'e' ':') WS* <UntilLineEnd?> Action12 LineEnd Action13 Step* OutlineExamples? Action14)> */
		nil,
		/* 5 OutlineExamples <- <(OS ('E' 'x' 'a' 'm' 'p' 'l' 'e' 's' ':') LineEnd Action15 Table? Action16)> */
		nil,
		/* 6 Step <- <(Tags <((&('B') ('B' 'u' 't')) | (&('O') ('O' 'r')) | (&('A') ('A' 'n' 'd')) | (&('T') ('T' 'h' 'e' 'n')) | (&('W') ('W' 'h' 'e' 'n')) | (&('G') ('G' 'i' 'v' 'e' 'n')))> Action17 WS* <UntilLineEnd> Action18 LineEnd Action19 StepArgument? Action20)> */
		func() bool {
			position92, tokenIndex92, depth92 := position, tokenIndex, depth
			{
				position93 := position
				depth++
				if !rules[ruleTags]() {
					goto l92
				}
				{
					position94 := position
					depth++
					{
						switch buffer[position] {
						case 'B':
							if buffer[position] != rune('B') {
								goto l92
							}
							position++
							if buffer[position] != rune('u') {
								goto l92
							}
							position++
							if buffer[position] != rune('t') {
								goto l92
							}
							position++
							break
						case 'O':
							if buffer[position] != rune('O') {
								goto l92
							}
							position++
							if buffer[position] != rune('r') {
								goto l92
							}
							position++
							break
						case 'A':
							if buffer[position] != rune('A') {
								goto l92
							}
							position++
							if buffer[position] != rune('n') {
								goto l92
							}
							position++
							if buffer[position] != rune('d') {
								goto l92
							}
							position++
							break
						case 'T':
							if buffer[position] != rune('T') {
								goto l92
							}
							position++
							if buffer[position] != rune('h') {
								goto l92
							}
							position++
							if buffer[position] != rune('e') {
								goto l92
							}
							position++
							if buffer[position] != rune('n') {
								goto l92
							}
							position++
							break
						case 'W':
							if buffer[position] != rune('W') {
								goto l92
							}
							position++
							if buffer[position] != rune('h') {
								goto l92
							}
							position++
							if buffer[position] != rune('e') {
								goto l92
							}
							position++
							if buffer[position] != rune('n') {
								goto l92
							}
							position++
							break
						default:
							if buffer[position] != rune('G') {
								goto l92
							}
							position++
							if buffer[position] != rune('i') {
								goto l92
							}
							position++
							if buffer[position] != rune('v') {
								goto l92
							}
							position++
							if buffer[position] != rune('e') {
								goto l92
							}
							position++
							if buffer[position] != rune('n') {
								goto l92
							}
							position++
							break
						}
					}

					depth--
					add(rulePegText, position94)
				}
				{
					add(ruleAction17, position)
				}
			l97:
				{
					position98, tokenIndex98, depth98 := position, tokenIndex, depth
					if !rules[ruleWS]() {
						goto l98
					}
					goto l97
				l98:
					position, tokenIndex, depth = position98, tokenIndex98, depth98
				}
				{
					position99 := position
					depth++
					if !rules[ruleUntilLineEnd]() {
						goto l92
					}
					depth--
					add(rulePegText, position99)
				}
				{
					add(ruleAction18, position)
				}
				if !rules[ruleLineEnd]() {
					goto l92
				}
				{
					add(ruleAction19, position)
				}
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					{
						position104 := position
						depth++
						{
							position105, tokenIndex105, depth105 := position, tokenIndex, depth
							if !rules[ruleTable]() {
								goto l106
							}
							goto l105
						l106:
							position, tokenIndex, depth = position105, tokenIndex105, depth105
							{
								position107 := position
								depth++
							l108:
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
									if !rules[ruleNL]() {
										goto l109
									}
									goto l108
								l109:
									position, tokenIndex, depth = position109, tokenIndex109, depth109
								}
								{
									position112 := position
									depth++
								l113:
									{
										position114, tokenIndex114, depth114 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l114
										}
										goto l113
									l114:
										position, tokenIndex, depth = position114, tokenIndex114, depth114
									}
									depth--
									add(rulePegText, position112)
								}
								if !rules[rulePyStringQuote]() {
									goto l102
								}
								if !rules[ruleNL]() {
									goto l102
								}
								{
									add(ruleAction21, position)
								}
							l116:
								{
									position117, tokenIndex117, depth117 := position, tokenIndex, depth
									{
										position118, tokenIndex118, depth118 := position, tokenIndex, depth
									l119:
										{
											position120, tokenIndex120, depth120 := position, tokenIndex, depth
											if !rules[ruleWS]() {
												goto l120
											}
											goto l119
										l120:
											position, tokenIndex, depth = position120, tokenIndex120, depth120
										}
										if !rules[rulePyStringQuote]() {
											goto l118
										}
										goto l117
									l118:
										position, tokenIndex, depth = position118, tokenIndex118, depth118
									}
									{
										position121 := position
										depth++
										{
											position122 := position
											depth++
											{
												position123 := position
												depth++
											l124:
												{
													position125, tokenIndex125, depth125 := position, tokenIndex, depth
													{
														position126, tokenIndex126, depth126 := position, tokenIndex, depth
														if buffer[position] != rune('\n') {
															goto l126
														}
														position++
														goto l125
													l126:
														position, tokenIndex, depth = position126, tokenIndex126, depth126
													}
													if !matchDot() {
														goto l125
													}
													goto l124
												l125:
													position, tokenIndex, depth = position125, tokenIndex125, depth125
												}
												depth--
												add(ruleUntilNL, position123)
											}
											if !rules[ruleNL]() {
												goto l117
											}
											depth--
											add(rulePegText, position122)
										}
										{
											add(ruleAction23, position)
										}
										depth--
										add(rulePyStringLine, position121)
									}
									goto l116
								l117:
									position, tokenIndex, depth = position117, tokenIndex117, depth117
								}
							l128:
								{
									position129, tokenIndex129, depth129 := position, tokenIndex, depth
									if !rules[ruleWS]() {
										goto l129
									}
									goto l128
								l129:
									position, tokenIndex, depth = position129, tokenIndex129, depth129
								}
								if !rules[rulePyStringQuote]() {
									goto l102
								}
								if !rules[ruleLineEnd]() {
									goto l102
								}
								{
									add(ruleAction22, position)
								}
								depth--
								add(rulePyString, position107)
							}
						}
					l105:
						depth--
						add(ruleStepArgument, position104)
					}
					goto l103
				l102:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
				}
			l103:
				{
					add(ruleAction20, position)
				}
				depth--
				add(ruleStep, position93)
			}
			return true
		l92:
			position, tokenIndex, depth = position92, tokenIndex92, depth92
			return false
		},
		/* 7 StepArgument <- <(Table / PyString)> */
		nil,
		/* 8 PyString <- <((WS* NL)* <WS*> PyStringQuote NL Action21 (!(WS* PyStringQuote) PyStringLine)* WS* PyStringQuote LineEnd Action22)> */
		nil,
		/* 9 PyStringQuote <- <('"' '"' '"')> */
		func() bool {
			position134, tokenIndex134, depth134 := position, tokenIndex, depth
			{
				position135 := position
				depth++
				if buffer[position] != rune('"') {
					goto l134
				}
				position++
				if buffer[position] != rune('"') {
					goto l134
				}
				position++
				if buffer[position] != rune('"') {
					goto l134
				}
				position++
				depth--
				add(rulePyStringQuote, position135)
			}
			return true
		l134:
			position, tokenIndex, depth = position134, tokenIndex134, depth134
			return false
		},
		/* 10 PyStringLine <- <(<(UntilNL NL)> Action23)> */
		nil,
		/* 11 Table <- <(Action24 TableRow+ Action25)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				{
					add(ruleAction24, position)
				}
				{
					position142 := position
					depth++
					{
						add(ruleAction26, position)
					}
					if !rules[ruleOS]() {
						goto l137
					}
					if buffer[position] != rune('|') {
						goto l137
					}
					position++
					{
						position146 := position
						depth++
						{
							position147 := position
							depth++
							{
								position150, tokenIndex150, depth150 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '|':
										if buffer[position] != rune('|') {
											goto l150
										}
										position++
										break
									case '\n':
										if buffer[position] != rune('\n') {
											goto l150
										}
										position++
										break
									default:
										if buffer[position] != rune('\r') {
											goto l150
										}
										position++
										break
									}
								}

								goto l137
							l150:
								position, tokenIndex, depth = position150, tokenIndex150, depth150
							}
							if !matchDot() {
								goto l137
							}
						l148:
							{
								position149, tokenIndex149, depth149 := position, tokenIndex, depth
								{
									position152, tokenIndex152, depth152 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l152
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l152
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l152
											}
											position++
											break
										}
									}

									goto l149
								l152:
									position, tokenIndex, depth = position152, tokenIndex152, depth152
								}
								if !matchDot() {
									goto l149
								}
								goto l148
							l149:
								position, tokenIndex, depth = position149, tokenIndex149, depth149
							}
							depth--
							add(rulePegText, position147)
						}
						if buffer[position] != rune('|') {
							goto l137
						}
						position++
						{
							add(ruleAction28, position)
						}
						depth--
						add(ruleTableCell, position146)
					}
				l144:
					{
						position145, tokenIndex145, depth145 := position, tokenIndex, depth
						{
							position155 := position
							depth++
							{
								position156 := position
								depth++
								{
									position159, tokenIndex159, depth159 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l159
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l159
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l159
											}
											position++
											break
										}
									}

									goto l145
								l159:
									position, tokenIndex, depth = position159, tokenIndex159, depth159
								}
								if !matchDot() {
									goto l145
								}
							l157:
								{
									position158, tokenIndex158, depth158 := position, tokenIndex, depth
									{
										position161, tokenIndex161, depth161 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l161
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l161
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l161
												}
												position++
												break
											}
										}

										goto l158
									l161:
										position, tokenIndex, depth = position161, tokenIndex161, depth161
									}
									if !matchDot() {
										goto l158
									}
									goto l157
								l158:
									position, tokenIndex, depth = position158, tokenIndex158, depth158
								}
								depth--
								add(rulePegText, position156)
							}
							if buffer[position] != rune('|') {
								goto l145
							}
							position++
							{
								add(ruleAction28, position)
							}
							depth--
							add(ruleTableCell, position155)
						}
						goto l144
					l145:
						position, tokenIndex, depth = position145, tokenIndex145, depth145
					}
					if !rules[ruleLineEnd]() {
						goto l137
					}
					{
						add(ruleAction27, position)
					}
					depth--
					add(ruleTableRow, position142)
				}
			l140:
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					{
						position165 := position
						depth++
						{
							add(ruleAction26, position)
						}
						if !rules[ruleOS]() {
							goto l141
						}
						if buffer[position] != rune('|') {
							goto l141
						}
						position++
						{
							position169 := position
							depth++
							{
								position170 := position
								depth++
								{
									position173, tokenIndex173, depth173 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l173
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l173
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l173
											}
											position++
											break
										}
									}

									goto l141
								l173:
									position, tokenIndex, depth = position173, tokenIndex173, depth173
								}
								if !matchDot() {
									goto l141
								}
							l171:
								{
									position172, tokenIndex172, depth172 := position, tokenIndex, depth
									{
										position175, tokenIndex175, depth175 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l175
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l175
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l175
												}
												position++
												break
											}
										}

										goto l172
									l175:
										position, tokenIndex, depth = position175, tokenIndex175, depth175
									}
									if !matchDot() {
										goto l172
									}
									goto l171
								l172:
									position, tokenIndex, depth = position172, tokenIndex172, depth172
								}
								depth--
								add(rulePegText, position170)
							}
							if buffer[position] != rune('|') {
								goto l141
							}
							position++
							{
								add(ruleAction28, position)
							}
							depth--
							add(ruleTableCell, position169)
						}
					l167:
						{
							position168, tokenIndex168, depth168 := position, tokenIndex, depth
							{
								position178 := position
								depth++
								{
									position179 := position
									depth++
									{
										position182, tokenIndex182, depth182 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l182
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l182
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l182
												}
												position++
												break
											}
										}

										goto l168
									l182:
										position, tokenIndex, depth = position182, tokenIndex182, depth182
									}
									if !matchDot() {
										goto l168
									}
								l180:
									{
										position181, tokenIndex181, depth181 := position, tokenIndex, depth
										{
											position184, tokenIndex184, depth184 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '|':
													if buffer[position] != rune('|') {
														goto l184
													}
													position++
													break
												case '\n':
													if buffer[position] != rune('\n') {
														goto l184
													}
													position++
													break
												default:
													if buffer[position] != rune('\r') {
														goto l184
													}
													position++
													break
												}
											}

											goto l181
										l184:
											position, tokenIndex, depth = position184, tokenIndex184, depth184
										}
										if !matchDot() {
											goto l181
										}
										goto l180
									l181:
										position, tokenIndex, depth = position181, tokenIndex181, depth181
									}
									depth--
									add(rulePegText, position179)
								}
								if buffer[position] != rune('|') {
									goto l168
								}
								position++
								{
									add(ruleAction28, position)
								}
								depth--
								add(ruleTableCell, position178)
							}
							goto l167
						l168:
							position, tokenIndex, depth = position168, tokenIndex168, depth168
						}
						if !rules[ruleLineEnd]() {
							goto l141
						}
						{
							add(ruleAction27, position)
						}
						depth--
						add(ruleTableRow, position165)
					}
					goto l140
				l141:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
				}
				{
					add(ruleAction25, position)
				}
				depth--
				add(ruleTable, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 12 TableRow <- <(Action26 OS '|' TableCell+ LineEnd Action27)> */
		nil,
		/* 13 TableCell <- <(<(!((&('|') '|') | (&('\n') '\n') | (&('\r') '\r')) .)+> '|' Action28)> */
		nil,
		/* 14 Tags <- <((Tag+ WS* LineEnd)* OS)> */
		func() bool {
			position191, tokenIndex191, depth191 := position, tokenIndex, depth
			{
				position192 := position
				depth++
			l193:
				{
					position194, tokenIndex194, depth194 := position, tokenIndex, depth
					{
						position197 := position
						depth++
						if !rules[ruleOS]() {
							goto l194
						}
						if buffer[position] != rune('@') {
							goto l194
						}
						position++
						{
							position198 := position
							depth++
							{
								position199 := position
								depth++
								{
									position202, tokenIndex202, depth202 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '#':
											if buffer[position] != rune('#') {
												goto l202
											}
											position++
											break
										case '"':
											if buffer[position] != rune('"') {
												goto l202
											}
											position++
											break
										case ' ':
											if buffer[position] != rune(' ') {
												goto l202
											}
											position++
											break
										case '\t':
											if buffer[position] != rune('\t') {
												goto l202
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l202
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l202
											}
											position++
											break
										}
									}

									goto l194
								l202:
									position, tokenIndex, depth = position202, tokenIndex202, depth202
								}
								if !matchDot() {
									goto l194
								}
							l200:
								{
									position201, tokenIndex201, depth201 := position, tokenIndex, depth
									{
										position204, tokenIndex204, depth204 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '#':
												if buffer[position] != rune('#') {
													goto l204
												}
												position++
												break
											case '"':
												if buffer[position] != rune('"') {
													goto l204
												}
												position++
												break
											case ' ':
												if buffer[position] != rune(' ') {
													goto l204
												}
												position++
												break
											case '\t':
												if buffer[position] != rune('\t') {
													goto l204
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l204
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l204
												}
												position++
												break
											}
										}

										goto l201
									l204:
										position, tokenIndex, depth = position204, tokenIndex204, depth204
									}
									if !matchDot() {
										goto l201
									}
									goto l200
								l201:
									position, tokenIndex, depth = position201, tokenIndex201, depth201
								}
								depth--
								add(ruleWord, position199)
							}
							depth--
							add(rulePegText, position198)
						}
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleTag, position197)
					}
				l195:
					{
						position196, tokenIndex196, depth196 := position, tokenIndex, depth
						{
							position207 := position
							depth++
							if !rules[ruleOS]() {
								goto l196
							}
							if buffer[position] != rune('@') {
								goto l196
							}
							position++
							{
								position208 := position
								depth++
								{
									position209 := position
									depth++
									{
										position212, tokenIndex212, depth212 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '#':
												if buffer[position] != rune('#') {
													goto l212
												}
												position++
												break
											case '"':
												if buffer[position] != rune('"') {
													goto l212
												}
												position++
												break
											case ' ':
												if buffer[position] != rune(' ') {
													goto l212
												}
												position++
												break
											case '\t':
												if buffer[position] != rune('\t') {
													goto l212
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l212
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l212
												}
												position++
												break
											}
										}

										goto l196
									l212:
										position, tokenIndex, depth = position212, tokenIndex212, depth212
									}
									if !matchDot() {
										goto l196
									}
								l210:
									{
										position211, tokenIndex211, depth211 := position, tokenIndex, depth
										{
											position214, tokenIndex214, depth214 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '#':
													if buffer[position] != rune('#') {
														goto l214
													}
													position++
													break
												case '"':
													if buffer[position] != rune('"') {
														goto l214
													}
													position++
													break
												case ' ':
													if buffer[position] != rune(' ') {
														goto l214
													}
													position++
													break
												case '\t':
													if buffer[position] != rune('\t') {
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

											goto l211
										l214:
											position, tokenIndex, depth = position214, tokenIndex214, depth214
										}
										if !matchDot() {
											goto l211
										}
										goto l210
									l211:
										position, tokenIndex, depth = position211, tokenIndex211, depth211
									}
									depth--
									add(ruleWord, position209)
								}
								depth--
								add(rulePegText, position208)
							}
							{
								add(ruleAction29, position)
							}
							depth--
							add(ruleTag, position207)
						}
						goto l195
					l196:
						position, tokenIndex, depth = position196, tokenIndex196, depth196
					}
				l217:
					{
						position218, tokenIndex218, depth218 := position, tokenIndex, depth
						if !rules[ruleWS]() {
							goto l218
						}
						goto l217
					l218:
						position, tokenIndex, depth = position218, tokenIndex218, depth218
					}
					if !rules[ruleLineEnd]() {
						goto l194
					}
					goto l193
				l194:
					position, tokenIndex, depth = position194, tokenIndex194, depth194
				}
				if !rules[ruleOS]() {
					goto l191
				}
				depth--
				add(ruleTags, position192)
			}
			return true
		l191:
			position, tokenIndex, depth = position191, tokenIndex191, depth191
			return false
		},
		/* 15 Tag <- <(OS '@' <Word> Action29)> */
		nil,
		/* 16 Word <- <(!((&('#') '#') | (&('"') '"') | (&(' ') ' ') | (&('\t') '\t') | (&('\n') '\n') | (&('\r') '\r')) .)+> */
		nil,
		/* 17 EscapedChar <- <('\\' .)> */
		func() bool {
			position221, tokenIndex221, depth221 := position, tokenIndex, depth
			{
				position222 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l221
				}
				position++
				if !matchDot() {
					goto l221
				}
				depth--
				add(ruleEscapedChar, position222)
			}
			return true
		l221:
			position, tokenIndex, depth = position221, tokenIndex221, depth221
			return false
		},
		/* 18 QuotedString <- <('"' (EscapedChar / (!((&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+)+ '"')> */
		nil,
		/* 19 UntilLineEnd <- <(EscapedChar / (!((&('#') '#') | (&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+ / QuotedString)+> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				{
					position228, tokenIndex228, depth228 := position, tokenIndex, depth
					if !rules[ruleEscapedChar]() {
						goto l229
					}
					goto l228
				l229:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
					{
						position233, tokenIndex233, depth233 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '#':
								if buffer[position] != rune('#') {
									goto l233
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l233
								}
								position++
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l233
								}
								position++
								break
							default:
								if buffer[position] != rune('\n') {
									goto l233
								}
								position++
								break
							}
						}

						goto l230
					l233:
						position, tokenIndex, depth = position233, tokenIndex233, depth233
					}
					if !matchDot() {
						goto l230
					}
				l231:
					{
						position232, tokenIndex232, depth232 := position, tokenIndex, depth
						{
							position235, tokenIndex235, depth235 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l235
									}
									position++
									break
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

							goto l232
						l235:
							position, tokenIndex, depth = position235, tokenIndex235, depth235
						}
						if !matchDot() {
							goto l232
						}
						goto l231
					l232:
						position, tokenIndex, depth = position232, tokenIndex232, depth232
					}
					goto l228
				l230:
					position, tokenIndex, depth = position228, tokenIndex228, depth228
					{
						position237 := position
						depth++
						if buffer[position] != rune('"') {
							goto l224
						}
						position++
						{
							position240, tokenIndex240, depth240 := position, tokenIndex, depth
							if !rules[ruleEscapedChar]() {
								goto l241
							}
							goto l240
						l241:
							position, tokenIndex, depth = position240, tokenIndex240, depth240
							{
								position244, tokenIndex244, depth244 := position, tokenIndex, depth
								{
									switch buffer[position] {
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

								goto l224
							l244:
								position, tokenIndex, depth = position244, tokenIndex244, depth244
							}
							if !matchDot() {
								goto l224
							}
						l242:
							{
								position243, tokenIndex243, depth243 := position, tokenIndex, depth
								{
									position246, tokenIndex246, depth246 := position, tokenIndex, depth
									{
										switch buffer[position] {
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
						}
					l240:
					l238:
						{
							position239, tokenIndex239, depth239 := position, tokenIndex, depth
							{
								position248, tokenIndex248, depth248 := position, tokenIndex, depth
								if !rules[ruleEscapedChar]() {
									goto l249
								}
								goto l248
							l249:
								position, tokenIndex, depth = position248, tokenIndex248, depth248
								{
									position252, tokenIndex252, depth252 := position, tokenIndex, depth
									{
										switch buffer[position] {
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

									goto l239
								l252:
									position, tokenIndex, depth = position252, tokenIndex252, depth252
								}
								if !matchDot() {
									goto l239
								}
							l250:
								{
									position251, tokenIndex251, depth251 := position, tokenIndex, depth
									{
										position254, tokenIndex254, depth254 := position, tokenIndex, depth
										{
											switch buffer[position] {
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
							}
						l248:
							goto l238
						l239:
							position, tokenIndex, depth = position239, tokenIndex239, depth239
						}
						if buffer[position] != rune('"') {
							goto l224
						}
						position++
						depth--
						add(ruleQuotedString, position237)
					}
				}
			l228:
			l226:
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					{
						position256, tokenIndex256, depth256 := position, tokenIndex, depth
						if !rules[ruleEscapedChar]() {
							goto l257
						}
						goto l256
					l257:
						position, tokenIndex, depth = position256, tokenIndex256, depth256
						{
							position261, tokenIndex261, depth261 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l261
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l261
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l261
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l261
									}
									position++
									break
								}
							}

							goto l258
						l261:
							position, tokenIndex, depth = position261, tokenIndex261, depth261
						}
						if !matchDot() {
							goto l258
						}
					l259:
						{
							position260, tokenIndex260, depth260 := position, tokenIndex, depth
							{
								position263, tokenIndex263, depth263 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '#':
										if buffer[position] != rune('#') {
											goto l263
										}
										position++
										break
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

								goto l260
							l263:
								position, tokenIndex, depth = position263, tokenIndex263, depth263
							}
							if !matchDot() {
								goto l260
							}
							goto l259
						l260:
							position, tokenIndex, depth = position260, tokenIndex260, depth260
						}
						goto l256
					l258:
						position, tokenIndex, depth = position256, tokenIndex256, depth256
						{
							position265 := position
							depth++
							if buffer[position] != rune('"') {
								goto l227
							}
							position++
							{
								position268, tokenIndex268, depth268 := position, tokenIndex, depth
								if !rules[ruleEscapedChar]() {
									goto l269
								}
								goto l268
							l269:
								position, tokenIndex, depth = position268, tokenIndex268, depth268
								{
									position272, tokenIndex272, depth272 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '"':
											if buffer[position] != rune('"') {
												goto l272
											}
											position++
											break
										case '\\':
											if buffer[position] != rune('\\') {
												goto l272
											}
											position++
											break
										default:
											if buffer[position] != rune('\n') {
												goto l272
											}
											position++
											break
										}
									}

									goto l227
								l272:
									position, tokenIndex, depth = position272, tokenIndex272, depth272
								}
								if !matchDot() {
									goto l227
								}
							l270:
								{
									position271, tokenIndex271, depth271 := position, tokenIndex, depth
									{
										position274, tokenIndex274, depth274 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l274
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l274
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l274
												}
												position++
												break
											}
										}

										goto l271
									l274:
										position, tokenIndex, depth = position274, tokenIndex274, depth274
									}
									if !matchDot() {
										goto l271
									}
									goto l270
								l271:
									position, tokenIndex, depth = position271, tokenIndex271, depth271
								}
							}
						l268:
						l266:
							{
								position267, tokenIndex267, depth267 := position, tokenIndex, depth
								{
									position276, tokenIndex276, depth276 := position, tokenIndex, depth
									if !rules[ruleEscapedChar]() {
										goto l277
									}
									goto l276
								l277:
									position, tokenIndex, depth = position276, tokenIndex276, depth276
									{
										position280, tokenIndex280, depth280 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l280
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l280
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l280
												}
												position++
												break
											}
										}

										goto l267
									l280:
										position, tokenIndex, depth = position280, tokenIndex280, depth280
									}
									if !matchDot() {
										goto l267
									}
								l278:
									{
										position279, tokenIndex279, depth279 := position, tokenIndex, depth
										{
											position282, tokenIndex282, depth282 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '"':
													if buffer[position] != rune('"') {
														goto l282
													}
													position++
													break
												case '\\':
													if buffer[position] != rune('\\') {
														goto l282
													}
													position++
													break
												default:
													if buffer[position] != rune('\n') {
														goto l282
													}
													position++
													break
												}
											}

											goto l279
										l282:
											position, tokenIndex, depth = position282, tokenIndex282, depth282
										}
										if !matchDot() {
											goto l279
										}
										goto l278
									l279:
										position, tokenIndex, depth = position279, tokenIndex279, depth279
									}
								}
							l276:
								goto l266
							l267:
								position, tokenIndex, depth = position267, tokenIndex267, depth267
							}
							if buffer[position] != rune('"') {
								goto l227
							}
							position++
							depth--
							add(ruleQuotedString, position265)
						}
					}
				l256:
					goto l226
				l227:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
				}
				depth--
				add(ruleUntilLineEnd, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 20 LineEnd <- <(WS* LineComment? (NL / !.))> */
		func() bool {
			position284, tokenIndex284, depth284 := position, tokenIndex, depth
			{
				position285 := position
				depth++
			l286:
				{
					position287, tokenIndex287, depth287 := position, tokenIndex, depth
					if !rules[ruleWS]() {
						goto l287
					}
					goto l286
				l287:
					position, tokenIndex, depth = position287, tokenIndex287, depth287
				}
				{
					position288, tokenIndex288, depth288 := position, tokenIndex, depth
					{
						position290 := position
						depth++
						if buffer[position] != rune('#') {
							goto l288
						}
						position++
						{
							position291 := position
							depth++
						l292:
							{
								position293, tokenIndex293, depth293 := position, tokenIndex, depth
								{
									position294, tokenIndex294, depth294 := position, tokenIndex, depth
									if buffer[position] != rune('\n') {
										goto l294
									}
									position++
									goto l293
								l294:
									position, tokenIndex, depth = position294, tokenIndex294, depth294
								}
								if !matchDot() {
									goto l293
								}
								goto l292
							l293:
								position, tokenIndex, depth = position293, tokenIndex293, depth293
							}
							depth--
							add(rulePegText, position291)
						}
						{
							add(ruleAction30, position)
						}
						depth--
						add(ruleLineComment, position290)
					}
					goto l289
				l288:
					position, tokenIndex, depth = position288, tokenIndex288, depth288
				}
			l289:
				{
					position296, tokenIndex296, depth296 := position, tokenIndex, depth
					if !rules[ruleNL]() {
						goto l297
					}
					goto l296
				l297:
					position, tokenIndex, depth = position296, tokenIndex296, depth296
					{
						position298, tokenIndex298, depth298 := position, tokenIndex, depth
						if !matchDot() {
							goto l298
						}
						goto l284
					l298:
						position, tokenIndex, depth = position298, tokenIndex298, depth298
					}
				}
			l296:
				depth--
				add(ruleLineEnd, position285)
			}
			return true
		l284:
			position, tokenIndex, depth = position284, tokenIndex284, depth284
			return false
		},
		/* 21 LineComment <- <('#' <(!'\n' .)*> Action30)> */
		nil,
		/* 22 BlankLine <- <(WS LineEnd)+> */
		nil,
		/* 23 OS <- <(NL / WS)*> */
		func() bool {
			{
				position302 := position
				depth++
			l303:
				{
					position304, tokenIndex304, depth304 := position, tokenIndex, depth
					{
						position305, tokenIndex305, depth305 := position, tokenIndex, depth
						if !rules[ruleNL]() {
							goto l306
						}
						goto l305
					l306:
						position, tokenIndex, depth = position305, tokenIndex305, depth305
						if !rules[ruleWS]() {
							goto l304
						}
					}
				l305:
					goto l303
				l304:
					position, tokenIndex, depth = position304, tokenIndex304, depth304
				}
				depth--
				add(ruleOS, position302)
			}
			return true
		},
		/* 24 WS <- <(' ' / '\t')> */
		func() bool {
			position307, tokenIndex307, depth307 := position, tokenIndex, depth
			{
				position308 := position
				depth++
				{
					position309, tokenIndex309, depth309 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l310
					}
					position++
					goto l309
				l310:
					position, tokenIndex, depth = position309, tokenIndex309, depth309
					if buffer[position] != rune('\t') {
						goto l307
					}
					position++
				}
			l309:
				depth--
				add(ruleWS, position308)
			}
			return true
		l307:
			position, tokenIndex, depth = position307, tokenIndex307, depth307
			return false
		},
		/* 25 UntilNL <- <(!'\n' .)*> */
		nil,
		/* 26 NL <- <('\n' / '\r' / ('\r' '\n'))> */
		func() bool {
			position312, tokenIndex312, depth312 := position, tokenIndex, depth
			{
				position313 := position
				depth++
				{
					position314, tokenIndex314, depth314 := position, tokenIndex, depth
					if buffer[position] != rune('\n') {
						goto l315
					}
					position++
					goto l314
				l315:
					position, tokenIndex, depth = position314, tokenIndex314, depth314
					if buffer[position] != rune('\r') {
						goto l316
					}
					position++
					goto l314
				l316:
					position, tokenIndex, depth = position314, tokenIndex314, depth314
					if buffer[position] != rune('\r') {
						goto l312
					}
					position++
					if buffer[position] != rune('\n') {
						goto l312
					}
					position++
				}
			l314:
				depth--
				add(ruleNL, position313)
			}
			return true
		l312:
			position, tokenIndex, depth = position312, tokenIndex312, depth312
			return false
		},
		nil,
		/* 29 Action0 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 30 Action1 <- <{ p.buf2 = buffer[begin:end] }> */
		nil,
		/* 31 Action2 <- <{ p.buf2 = p.buf2 + buffer[begin:end] }> */
		nil,
		/* 32 Action3 <- <{p.buf2 = p.buf2 + "\n" }> */
		nil,
		/* 33 Action4 <- <{ p.beginFeature(trimWS(p.buf1), trimWSML(p.buf2), p.buftags); p.buftags = nil }> */
		nil,
		/* 34 Action5 <- <{ p.endFeature() }> */
		nil,
		/* 35 Action6 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 36 Action7 <- <{ p.beginBackground(p.buf1, p.buftags); p.buftags = nil }> */
		nil,
		/* 37 Action8 <- <{ p.endBackground() }> */
		nil,
		/* 38 Action9 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 39 Action10 <- <{ p.beginScenario(p.buf1, p.buftags); p.buftags = nil }> */
		nil,
		/* 40 Action11 <- <{ p.endScenario() }> */
		nil,
		/* 41 Action12 <- <{ p.buf1 = buffer[begin:end] }> */
		nil,
		/* 42 Action13 <- <{ p.beginOutline(p.buf1, p.buftags); p.buftags = nil }> */
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
		/* 48 Action19 <- <{ p.beginStep(p.buf1, p.buf2) }> */
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
		/* 57 Action28 <- <{ p.beginTableCell(); p.endTableCell(buffer[begin:end]) }> */
		nil,
		/* 58 Action29 <- <{ p.buftags = append(p.buftags, buffer[begin:end]) }> */
		nil,
		/* 59 Action30 <- <{ p.bufcmt = buffer[begin:end] }> */
		nil,
	}
	p.rules = rules
}
