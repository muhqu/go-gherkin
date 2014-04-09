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
			p.beginFeature(trimWS(trimWS(p.buf1)), trimWSML(p.buf2), p.buftags)
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
										{
											position50, tokenIndex50, depth50 := position, tokenIndex, depth
											if !rules[ruleStep]() {
												goto l51
											}
											goto l50
										l51:
											position, tokenIndex, depth = position50, tokenIndex50, depth50
											if !rules[ruleBlankLine]() {
												goto l49
											}
										}
									l50:
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
									position54 := position
									depth++
									if !rules[ruleTags]() {
										goto l53
									}
									if buffer[position] != rune('S') {
										goto l53
									}
									position++
									if buffer[position] != rune('c') {
										goto l53
									}
									position++
									if buffer[position] != rune('e') {
										goto l53
									}
									position++
									if buffer[position] != rune('n') {
										goto l53
									}
									position++
									if buffer[position] != rune('a') {
										goto l53
									}
									position++
									if buffer[position] != rune('r') {
										goto l53
									}
									position++
									if buffer[position] != rune('i') {
										goto l53
									}
									position++
									if buffer[position] != rune('o') {
										goto l53
									}
									position++
									if buffer[position] != rune(':') {
										goto l53
									}
									position++
								l55:
									{
										position56, tokenIndex56, depth56 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l56
										}
										goto l55
									l56:
										position, tokenIndex, depth = position56, tokenIndex56, depth56
									}
									{
										position57 := position
										depth++
										{
											position58, tokenIndex58, depth58 := position, tokenIndex, depth
											if !rules[ruleUntilLineEnd]() {
												goto l58
											}
											goto l59
										l58:
											position, tokenIndex, depth = position58, tokenIndex58, depth58
										}
									l59:
										depth--
										add(rulePegText, position57)
									}
									{
										add(ruleAction9, position)
									}
									if !rules[ruleLineEnd]() {
										goto l53
									}
									{
										add(ruleAction10, position)
									}
								l62:
									{
										position63, tokenIndex63, depth63 := position, tokenIndex, depth
										{
											position64, tokenIndex64, depth64 := position, tokenIndex, depth
											if !rules[ruleStep]() {
												goto l65
											}
											goto l64
										l65:
											position, tokenIndex, depth = position64, tokenIndex64, depth64
											if !rules[ruleBlankLine]() {
												goto l63
											}
										}
									l64:
										goto l62
									l63:
										position, tokenIndex, depth = position63, tokenIndex63, depth63
									}
									{
										add(ruleAction11, position)
									}
									depth--
									add(ruleScenario, position54)
								}
								goto l38
							l53:
								position, tokenIndex, depth = position38, tokenIndex38, depth38
								{
									position68 := position
									depth++
									if !rules[ruleTags]() {
										goto l67
									}
									if buffer[position] != rune('S') {
										goto l67
									}
									position++
									if buffer[position] != rune('c') {
										goto l67
									}
									position++
									if buffer[position] != rune('e') {
										goto l67
									}
									position++
									if buffer[position] != rune('n') {
										goto l67
									}
									position++
									if buffer[position] != rune('a') {
										goto l67
									}
									position++
									if buffer[position] != rune('r') {
										goto l67
									}
									position++
									if buffer[position] != rune('i') {
										goto l67
									}
									position++
									if buffer[position] != rune('o') {
										goto l67
									}
									position++
									if buffer[position] != rune(' ') {
										goto l67
									}
									position++
									if buffer[position] != rune('O') {
										goto l67
									}
									position++
									if buffer[position] != rune('u') {
										goto l67
									}
									position++
									if buffer[position] != rune('t') {
										goto l67
									}
									position++
									if buffer[position] != rune('l') {
										goto l67
									}
									position++
									if buffer[position] != rune('i') {
										goto l67
									}
									position++
									if buffer[position] != rune('n') {
										goto l67
									}
									position++
									if buffer[position] != rune('e') {
										goto l67
									}
									position++
									if buffer[position] != rune(':') {
										goto l67
									}
									position++
								l69:
									{
										position70, tokenIndex70, depth70 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l70
										}
										goto l69
									l70:
										position, tokenIndex, depth = position70, tokenIndex70, depth70
									}
									{
										position71 := position
										depth++
										{
											position72, tokenIndex72, depth72 := position, tokenIndex, depth
											if !rules[ruleUntilLineEnd]() {
												goto l72
											}
											goto l73
										l72:
											position, tokenIndex, depth = position72, tokenIndex72, depth72
										}
									l73:
										depth--
										add(rulePegText, position71)
									}
									{
										add(ruleAction12, position)
									}
									if !rules[ruleLineEnd]() {
										goto l67
									}
									{
										add(ruleAction13, position)
									}
								l76:
									{
										position77, tokenIndex77, depth77 := position, tokenIndex, depth
										{
											position78, tokenIndex78, depth78 := position, tokenIndex, depth
											if !rules[ruleStep]() {
												goto l79
											}
											goto l78
										l79:
											position, tokenIndex, depth = position78, tokenIndex78, depth78
											if !rules[ruleBlankLine]() {
												goto l77
											}
										}
									l78:
										goto l76
									l77:
										position, tokenIndex, depth = position77, tokenIndex77, depth77
									}
									{
										position80, tokenIndex80, depth80 := position, tokenIndex, depth
										{
											position82 := position
											depth++
											if !rules[ruleOS]() {
												goto l80
											}
											if buffer[position] != rune('E') {
												goto l80
											}
											position++
											if buffer[position] != rune('x') {
												goto l80
											}
											position++
											if buffer[position] != rune('a') {
												goto l80
											}
											position++
											if buffer[position] != rune('m') {
												goto l80
											}
											position++
											if buffer[position] != rune('p') {
												goto l80
											}
											position++
											if buffer[position] != rune('l') {
												goto l80
											}
											position++
											if buffer[position] != rune('e') {
												goto l80
											}
											position++
											if buffer[position] != rune('s') {
												goto l80
											}
											position++
											if buffer[position] != rune(':') {
												goto l80
											}
											position++
											if !rules[ruleLineEnd]() {
												goto l80
											}
											{
												add(ruleAction15, position)
											}
											{
												position84, tokenIndex84, depth84 := position, tokenIndex, depth
												if !rules[ruleTable]() {
													goto l84
												}
												goto l85
											l84:
												position, tokenIndex, depth = position84, tokenIndex84, depth84
											}
										l85:
											{
												add(ruleAction16, position)
											}
											depth--
											add(ruleOutlineExamples, position82)
										}
										goto l81
									l80:
										position, tokenIndex, depth = position80, tokenIndex80, depth80
									}
								l81:
									{
										add(ruleAction14, position)
									}
									depth--
									add(ruleOutline, position68)
								}
								goto l38
							l67:
								position, tokenIndex, depth = position38, tokenIndex38, depth38
								if !rules[ruleBlankLine]() {
									goto l37
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
					position89, tokenIndex89, depth89 := position, tokenIndex, depth
					if !matchDot() {
						goto l89
					}
					goto l0
				l89:
					position, tokenIndex, depth = position89, tokenIndex89, depth89
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
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
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
					{
						switch buffer[position] {
						case 'B':
							if buffer[position] != rune('B') {
								goto l95
							}
							position++
							if buffer[position] != rune('u') {
								goto l95
							}
							position++
							if buffer[position] != rune('t') {
								goto l95
							}
							position++
							break
						case 'O':
							if buffer[position] != rune('O') {
								goto l95
							}
							position++
							if buffer[position] != rune('r') {
								goto l95
							}
							position++
							break
						case 'A':
							if buffer[position] != rune('A') {
								goto l95
							}
							position++
							if buffer[position] != rune('n') {
								goto l95
							}
							position++
							if buffer[position] != rune('d') {
								goto l95
							}
							position++
							break
						case 'T':
							if buffer[position] != rune('T') {
								goto l95
							}
							position++
							if buffer[position] != rune('h') {
								goto l95
							}
							position++
							if buffer[position] != rune('e') {
								goto l95
							}
							position++
							if buffer[position] != rune('n') {
								goto l95
							}
							position++
							break
						case 'W':
							if buffer[position] != rune('W') {
								goto l95
							}
							position++
							if buffer[position] != rune('h') {
								goto l95
							}
							position++
							if buffer[position] != rune('e') {
								goto l95
							}
							position++
							if buffer[position] != rune('n') {
								goto l95
							}
							position++
							break
						default:
							if buffer[position] != rune('G') {
								goto l95
							}
							position++
							if buffer[position] != rune('i') {
								goto l95
							}
							position++
							if buffer[position] != rune('v') {
								goto l95
							}
							position++
							if buffer[position] != rune('e') {
								goto l95
							}
							position++
							if buffer[position] != rune('n') {
								goto l95
							}
							position++
							break
						}
					}

					depth--
					add(rulePegText, position99)
				}
				{
					add(ruleAction17, position)
				}
			l102:
				{
					position103, tokenIndex103, depth103 := position, tokenIndex, depth
					if !rules[ruleWS]() {
						goto l103
					}
					goto l102
				l103:
					position, tokenIndex, depth = position103, tokenIndex103, depth103
				}
				{
					position104 := position
					depth++
					if !rules[ruleUntilLineEnd]() {
						goto l95
					}
					depth--
					add(rulePegText, position104)
				}
				{
					add(ruleAction18, position)
				}
				if !rules[ruleLineEnd]() {
					goto l95
				}
				{
					add(ruleAction19, position)
				}
				{
					position107, tokenIndex107, depth107 := position, tokenIndex, depth
					{
						position109 := position
						depth++
						{
							position110, tokenIndex110, depth110 := position, tokenIndex, depth
							if !rules[ruleTable]() {
								goto l111
							}
							goto l110
						l111:
							position, tokenIndex, depth = position110, tokenIndex110, depth110
							{
								position112 := position
								depth++
							l113:
								{
									position114, tokenIndex114, depth114 := position, tokenIndex, depth
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
									if !rules[ruleNL]() {
										goto l114
									}
									goto l113
								l114:
									position, tokenIndex, depth = position114, tokenIndex114, depth114
								}
								{
									position117 := position
									depth++
								l118:
									{
										position119, tokenIndex119, depth119 := position, tokenIndex, depth
										if !rules[ruleWS]() {
											goto l119
										}
										goto l118
									l119:
										position, tokenIndex, depth = position119, tokenIndex119, depth119
									}
									depth--
									add(rulePegText, position117)
								}
								if !rules[rulePyStringQuote]() {
									goto l107
								}
								if !rules[ruleNL]() {
									goto l107
								}
								{
									add(ruleAction21, position)
								}
							l121:
								{
									position122, tokenIndex122, depth122 := position, tokenIndex, depth
									{
										position123, tokenIndex123, depth123 := position, tokenIndex, depth
									l124:
										{
											position125, tokenIndex125, depth125 := position, tokenIndex, depth
											if !rules[ruleWS]() {
												goto l125
											}
											goto l124
										l125:
											position, tokenIndex, depth = position125, tokenIndex125, depth125
										}
										if !rules[rulePyStringQuote]() {
											goto l123
										}
										goto l122
									l123:
										position, tokenIndex, depth = position123, tokenIndex123, depth123
									}
									{
										position126 := position
										depth++
										{
											position127 := position
											depth++
											{
												position128 := position
												depth++
											l129:
												{
													position130, tokenIndex130, depth130 := position, tokenIndex, depth
													{
														position131, tokenIndex131, depth131 := position, tokenIndex, depth
														if buffer[position] != rune('\n') {
															goto l131
														}
														position++
														goto l130
													l131:
														position, tokenIndex, depth = position131, tokenIndex131, depth131
													}
													if !matchDot() {
														goto l130
													}
													goto l129
												l130:
													position, tokenIndex, depth = position130, tokenIndex130, depth130
												}
												depth--
												add(ruleUntilNL, position128)
											}
											depth--
											add(rulePegText, position127)
										}
										if !rules[ruleNL]() {
											goto l122
										}
										{
											add(ruleAction23, position)
										}
										depth--
										add(rulePyStringLine, position126)
									}
									goto l121
								l122:
									position, tokenIndex, depth = position122, tokenIndex122, depth122
								}
							l133:
								{
									position134, tokenIndex134, depth134 := position, tokenIndex, depth
									if !rules[ruleWS]() {
										goto l134
									}
									goto l133
								l134:
									position, tokenIndex, depth = position134, tokenIndex134, depth134
								}
								if !rules[rulePyStringQuote]() {
									goto l107
								}
								if !rules[ruleLineEnd]() {
									goto l107
								}
								{
									add(ruleAction22, position)
								}
								depth--
								add(rulePyString, position112)
							}
						}
					l110:
						depth--
						add(ruleStepArgument, position109)
					}
					goto l108
				l107:
					position, tokenIndex, depth = position107, tokenIndex107, depth107
				}
			l108:
				{
					add(ruleAction20, position)
				}
				depth--
				add(ruleStep, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 7 StepArgument <- <(Table / PyString)> */
		nil,
		/* 8 PyString <- <((WS* NL)* <WS*> PyStringQuote NL Action21 (!(WS* PyStringQuote) PyStringLine)* WS* PyStringQuote LineEnd Action22)> */
		nil,
		/* 9 PyStringQuote <- <('"' '"' '"')> */
		func() bool {
			position139, tokenIndex139, depth139 := position, tokenIndex, depth
			{
				position140 := position
				depth++
				if buffer[position] != rune('"') {
					goto l139
				}
				position++
				if buffer[position] != rune('"') {
					goto l139
				}
				position++
				if buffer[position] != rune('"') {
					goto l139
				}
				position++
				depth--
				add(rulePyStringQuote, position140)
			}
			return true
		l139:
			position, tokenIndex, depth = position139, tokenIndex139, depth139
			return false
		},
		/* 10 PyStringLine <- <(<UntilNL> NL Action23)> */
		nil,
		/* 11 Table <- <(Action24 TableRow+ Action25)> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				{
					add(ruleAction24, position)
				}
				{
					position147 := position
					depth++
					{
						add(ruleAction26, position)
					}
					if !rules[ruleOS]() {
						goto l142
					}
					if buffer[position] != rune('|') {
						goto l142
					}
					position++
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

								goto l142
							l155:
								position, tokenIndex, depth = position155, tokenIndex155, depth155
							}
							if !matchDot() {
								goto l142
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
							goto l142
						}
						position++
						{
							add(ruleAction28, position)
						}
						depth--
						add(ruleTableCell, position151)
					}
				l149:
					{
						position150, tokenIndex150, depth150 := position, tokenIndex, depth
						{
							position160 := position
							depth++
							{
								position161 := position
								depth++
								{
									position164, tokenIndex164, depth164 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '|':
											if buffer[position] != rune('|') {
												goto l164
											}
											position++
											break
										case '\n':
											if buffer[position] != rune('\n') {
												goto l164
											}
											position++
											break
										default:
											if buffer[position] != rune('\r') {
												goto l164
											}
											position++
											break
										}
									}

									goto l150
								l164:
									position, tokenIndex, depth = position164, tokenIndex164, depth164
								}
								if !matchDot() {
									goto l150
								}
							l162:
								{
									position163, tokenIndex163, depth163 := position, tokenIndex, depth
									{
										position166, tokenIndex166, depth166 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l166
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l166
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l166
												}
												position++
												break
											}
										}

										goto l163
									l166:
										position, tokenIndex, depth = position166, tokenIndex166, depth166
									}
									if !matchDot() {
										goto l163
									}
									goto l162
								l163:
									position, tokenIndex, depth = position163, tokenIndex163, depth163
								}
								depth--
								add(rulePegText, position161)
							}
							if buffer[position] != rune('|') {
								goto l150
							}
							position++
							{
								add(ruleAction28, position)
							}
							depth--
							add(ruleTableCell, position160)
						}
						goto l149
					l150:
						position, tokenIndex, depth = position150, tokenIndex150, depth150
					}
					if !rules[ruleLineEnd]() {
						goto l142
					}
					{
						add(ruleAction27, position)
					}
					depth--
					add(ruleTableRow, position147)
				}
			l145:
				{
					position146, tokenIndex146, depth146 := position, tokenIndex, depth
					{
						position170 := position
						depth++
						{
							add(ruleAction26, position)
						}
						if !rules[ruleOS]() {
							goto l146
						}
						if buffer[position] != rune('|') {
							goto l146
						}
						position++
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

									goto l146
								l178:
									position, tokenIndex, depth = position178, tokenIndex178, depth178
								}
								if !matchDot() {
									goto l146
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
								goto l146
							}
							position++
							{
								add(ruleAction28, position)
							}
							depth--
							add(ruleTableCell, position174)
						}
					l172:
						{
							position173, tokenIndex173, depth173 := position, tokenIndex, depth
							{
								position183 := position
								depth++
								{
									position184 := position
									depth++
									{
										position187, tokenIndex187, depth187 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '|':
												if buffer[position] != rune('|') {
													goto l187
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l187
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l187
												}
												position++
												break
											}
										}

										goto l173
									l187:
										position, tokenIndex, depth = position187, tokenIndex187, depth187
									}
									if !matchDot() {
										goto l173
									}
								l185:
									{
										position186, tokenIndex186, depth186 := position, tokenIndex, depth
										{
											position189, tokenIndex189, depth189 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '|':
													if buffer[position] != rune('|') {
														goto l189
													}
													position++
													break
												case '\n':
													if buffer[position] != rune('\n') {
														goto l189
													}
													position++
													break
												default:
													if buffer[position] != rune('\r') {
														goto l189
													}
													position++
													break
												}
											}

											goto l186
										l189:
											position, tokenIndex, depth = position189, tokenIndex189, depth189
										}
										if !matchDot() {
											goto l186
										}
										goto l185
									l186:
										position, tokenIndex, depth = position186, tokenIndex186, depth186
									}
									depth--
									add(rulePegText, position184)
								}
								if buffer[position] != rune('|') {
									goto l173
								}
								position++
								{
									add(ruleAction28, position)
								}
								depth--
								add(ruleTableCell, position183)
							}
							goto l172
						l173:
							position, tokenIndex, depth = position173, tokenIndex173, depth173
						}
						if !rules[ruleLineEnd]() {
							goto l146
						}
						{
							add(ruleAction27, position)
						}
						depth--
						add(ruleTableRow, position170)
					}
					goto l145
				l146:
					position, tokenIndex, depth = position146, tokenIndex146, depth146
				}
				{
					add(ruleAction25, position)
				}
				depth--
				add(ruleTable, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
		/* 12 TableRow <- <(Action26 OS '|' TableCell+ LineEnd Action27)> */
		nil,
		/* 13 TableCell <- <(<(!((&('|') '|') | (&('\n') '\n') | (&('\r') '\r')) .)+> '|' Action28)> */
		nil,
		/* 14 Tags <- <((Tag+ WS* LineEnd?)* OS)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
			l198:
				{
					position199, tokenIndex199, depth199 := position, tokenIndex, depth
					{
						position202 := position
						depth++
						if !rules[ruleOS]() {
							goto l199
						}
						if buffer[position] != rune('@') {
							goto l199
						}
						position++
						{
							position203 := position
							depth++
							{
								position204 := position
								depth++
								{
									position207, tokenIndex207, depth207 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '#':
											if buffer[position] != rune('#') {
												goto l207
											}
											position++
											break
										case '"':
											if buffer[position] != rune('"') {
												goto l207
											}
											position++
											break
										case ' ':
											if buffer[position] != rune(' ') {
												goto l207
											}
											position++
											break
										case '\t':
											if buffer[position] != rune('\t') {
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

									goto l199
								l207:
									position, tokenIndex, depth = position207, tokenIndex207, depth207
								}
								if !matchDot() {
									goto l199
								}
							l205:
								{
									position206, tokenIndex206, depth206 := position, tokenIndex, depth
									{
										position209, tokenIndex209, depth209 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '#':
												if buffer[position] != rune('#') {
													goto l209
												}
												position++
												break
											case '"':
												if buffer[position] != rune('"') {
													goto l209
												}
												position++
												break
											case ' ':
												if buffer[position] != rune(' ') {
													goto l209
												}
												position++
												break
											case '\t':
												if buffer[position] != rune('\t') {
													goto l209
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l209
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l209
												}
												position++
												break
											}
										}

										goto l206
									l209:
										position, tokenIndex, depth = position209, tokenIndex209, depth209
									}
									if !matchDot() {
										goto l206
									}
									goto l205
								l206:
									position, tokenIndex, depth = position206, tokenIndex206, depth206
								}
								depth--
								add(ruleWord, position204)
							}
							depth--
							add(rulePegText, position203)
						}
						{
							add(ruleAction29, position)
						}
						depth--
						add(ruleTag, position202)
					}
				l200:
					{
						position201, tokenIndex201, depth201 := position, tokenIndex, depth
						{
							position212 := position
							depth++
							if !rules[ruleOS]() {
								goto l201
							}
							if buffer[position] != rune('@') {
								goto l201
							}
							position++
							{
								position213 := position
								depth++
								{
									position214 := position
									depth++
									{
										position217, tokenIndex217, depth217 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '#':
												if buffer[position] != rune('#') {
													goto l217
												}
												position++
												break
											case '"':
												if buffer[position] != rune('"') {
													goto l217
												}
												position++
												break
											case ' ':
												if buffer[position] != rune(' ') {
													goto l217
												}
												position++
												break
											case '\t':
												if buffer[position] != rune('\t') {
													goto l217
												}
												position++
												break
											case '\n':
												if buffer[position] != rune('\n') {
													goto l217
												}
												position++
												break
											default:
												if buffer[position] != rune('\r') {
													goto l217
												}
												position++
												break
											}
										}

										goto l201
									l217:
										position, tokenIndex, depth = position217, tokenIndex217, depth217
									}
									if !matchDot() {
										goto l201
									}
								l215:
									{
										position216, tokenIndex216, depth216 := position, tokenIndex, depth
										{
											position219, tokenIndex219, depth219 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '#':
													if buffer[position] != rune('#') {
														goto l219
													}
													position++
													break
												case '"':
													if buffer[position] != rune('"') {
														goto l219
													}
													position++
													break
												case ' ':
													if buffer[position] != rune(' ') {
														goto l219
													}
													position++
													break
												case '\t':
													if buffer[position] != rune('\t') {
														goto l219
													}
													position++
													break
												case '\n':
													if buffer[position] != rune('\n') {
														goto l219
													}
													position++
													break
												default:
													if buffer[position] != rune('\r') {
														goto l219
													}
													position++
													break
												}
											}

											goto l216
										l219:
											position, tokenIndex, depth = position219, tokenIndex219, depth219
										}
										if !matchDot() {
											goto l216
										}
										goto l215
									l216:
										position, tokenIndex, depth = position216, tokenIndex216, depth216
									}
									depth--
									add(ruleWord, position214)
								}
								depth--
								add(rulePegText, position213)
							}
							{
								add(ruleAction29, position)
							}
							depth--
							add(ruleTag, position212)
						}
						goto l200
					l201:
						position, tokenIndex, depth = position201, tokenIndex201, depth201
					}
				l222:
					{
						position223, tokenIndex223, depth223 := position, tokenIndex, depth
						if !rules[ruleWS]() {
							goto l223
						}
						goto l222
					l223:
						position, tokenIndex, depth = position223, tokenIndex223, depth223
					}
					{
						position224, tokenIndex224, depth224 := position, tokenIndex, depth
						if !rules[ruleLineEnd]() {
							goto l224
						}
						goto l225
					l224:
						position, tokenIndex, depth = position224, tokenIndex224, depth224
					}
				l225:
					goto l198
				l199:
					position, tokenIndex, depth = position199, tokenIndex199, depth199
				}
				if !rules[ruleOS]() {
					goto l196
				}
				depth--
				add(ruleTags, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 15 Tag <- <(OS '@' <Word> Action29)> */
		nil,
		/* 16 Word <- <(!((&('#') '#') | (&('"') '"') | (&(' ') ' ') | (&('\t') '\t') | (&('\n') '\n') | (&('\r') '\r')) .)+> */
		nil,
		/* 17 EscapedChar <- <('\\' .)> */
		func() bool {
			position228, tokenIndex228, depth228 := position, tokenIndex, depth
			{
				position229 := position
				depth++
				if buffer[position] != rune('\\') {
					goto l228
				}
				position++
				if !matchDot() {
					goto l228
				}
				depth--
				add(ruleEscapedChar, position229)
			}
			return true
		l228:
			position, tokenIndex, depth = position228, tokenIndex228, depth228
			return false
		},
		/* 18 QuotedString <- <('"' (EscapedChar / (!((&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+)+ '"')> */
		nil,
		/* 19 UntilLineEnd <- <(EscapedChar / (!((&('#') '#') | (&('"') '"') | (&('\\') '\\') | (&('\n') '\n')) .)+ / QuotedString)+> */
		func() bool {
			position231, tokenIndex231, depth231 := position, tokenIndex, depth
			{
				position232 := position
				depth++
				{
					position235, tokenIndex235, depth235 := position, tokenIndex, depth
					if !rules[ruleEscapedChar]() {
						goto l236
					}
					goto l235
				l236:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
					{
						position240, tokenIndex240, depth240 := position, tokenIndex, depth
						{
							switch buffer[position] {
							case '#':
								if buffer[position] != rune('#') {
									goto l240
								}
								position++
								break
							case '"':
								if buffer[position] != rune('"') {
									goto l240
								}
								position++
								break
							case '\\':
								if buffer[position] != rune('\\') {
									goto l240
								}
								position++
								break
							default:
								if buffer[position] != rune('\n') {
									goto l240
								}
								position++
								break
							}
						}

						goto l237
					l240:
						position, tokenIndex, depth = position240, tokenIndex240, depth240
					}
					if !matchDot() {
						goto l237
					}
				l238:
					{
						position239, tokenIndex239, depth239 := position, tokenIndex, depth
						{
							position242, tokenIndex242, depth242 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l242
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l242
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l242
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l242
									}
									position++
									break
								}
							}

							goto l239
						l242:
							position, tokenIndex, depth = position242, tokenIndex242, depth242
						}
						if !matchDot() {
							goto l239
						}
						goto l238
					l239:
						position, tokenIndex, depth = position239, tokenIndex239, depth239
					}
					goto l235
				l237:
					position, tokenIndex, depth = position235, tokenIndex235, depth235
					{
						position244 := position
						depth++
						if buffer[position] != rune('"') {
							goto l231
						}
						position++
						{
							position247, tokenIndex247, depth247 := position, tokenIndex, depth
							if !rules[ruleEscapedChar]() {
								goto l248
							}
							goto l247
						l248:
							position, tokenIndex, depth = position247, tokenIndex247, depth247
							{
								position251, tokenIndex251, depth251 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '"':
										if buffer[position] != rune('"') {
											goto l251
										}
										position++
										break
									case '\\':
										if buffer[position] != rune('\\') {
											goto l251
										}
										position++
										break
									default:
										if buffer[position] != rune('\n') {
											goto l251
										}
										position++
										break
									}
								}

								goto l231
							l251:
								position, tokenIndex, depth = position251, tokenIndex251, depth251
							}
							if !matchDot() {
								goto l231
							}
						l249:
							{
								position250, tokenIndex250, depth250 := position, tokenIndex, depth
								{
									position253, tokenIndex253, depth253 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '"':
											if buffer[position] != rune('"') {
												goto l253
											}
											position++
											break
										case '\\':
											if buffer[position] != rune('\\') {
												goto l253
											}
											position++
											break
										default:
											if buffer[position] != rune('\n') {
												goto l253
											}
											position++
											break
										}
									}

									goto l250
								l253:
									position, tokenIndex, depth = position253, tokenIndex253, depth253
								}
								if !matchDot() {
									goto l250
								}
								goto l249
							l250:
								position, tokenIndex, depth = position250, tokenIndex250, depth250
							}
						}
					l247:
					l245:
						{
							position246, tokenIndex246, depth246 := position, tokenIndex, depth
							{
								position255, tokenIndex255, depth255 := position, tokenIndex, depth
								if !rules[ruleEscapedChar]() {
									goto l256
								}
								goto l255
							l256:
								position, tokenIndex, depth = position255, tokenIndex255, depth255
								{
									position259, tokenIndex259, depth259 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '"':
											if buffer[position] != rune('"') {
												goto l259
											}
											position++
											break
										case '\\':
											if buffer[position] != rune('\\') {
												goto l259
											}
											position++
											break
										default:
											if buffer[position] != rune('\n') {
												goto l259
											}
											position++
											break
										}
									}

									goto l246
								l259:
									position, tokenIndex, depth = position259, tokenIndex259, depth259
								}
								if !matchDot() {
									goto l246
								}
							l257:
								{
									position258, tokenIndex258, depth258 := position, tokenIndex, depth
									{
										position261, tokenIndex261, depth261 := position, tokenIndex, depth
										{
											switch buffer[position] {
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
									goto l257
								l258:
									position, tokenIndex, depth = position258, tokenIndex258, depth258
								}
							}
						l255:
							goto l245
						l246:
							position, tokenIndex, depth = position246, tokenIndex246, depth246
						}
						if buffer[position] != rune('"') {
							goto l231
						}
						position++
						depth--
						add(ruleQuotedString, position244)
					}
				}
			l235:
			l233:
				{
					position234, tokenIndex234, depth234 := position, tokenIndex, depth
					{
						position263, tokenIndex263, depth263 := position, tokenIndex, depth
						if !rules[ruleEscapedChar]() {
							goto l264
						}
						goto l263
					l264:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
						{
							position268, tokenIndex268, depth268 := position, tokenIndex, depth
							{
								switch buffer[position] {
								case '#':
									if buffer[position] != rune('#') {
										goto l268
									}
									position++
									break
								case '"':
									if buffer[position] != rune('"') {
										goto l268
									}
									position++
									break
								case '\\':
									if buffer[position] != rune('\\') {
										goto l268
									}
									position++
									break
								default:
									if buffer[position] != rune('\n') {
										goto l268
									}
									position++
									break
								}
							}

							goto l265
						l268:
							position, tokenIndex, depth = position268, tokenIndex268, depth268
						}
						if !matchDot() {
							goto l265
						}
					l266:
						{
							position267, tokenIndex267, depth267 := position, tokenIndex, depth
							{
								position270, tokenIndex270, depth270 := position, tokenIndex, depth
								{
									switch buffer[position] {
									case '#':
										if buffer[position] != rune('#') {
											goto l270
										}
										position++
										break
									case '"':
										if buffer[position] != rune('"') {
											goto l270
										}
										position++
										break
									case '\\':
										if buffer[position] != rune('\\') {
											goto l270
										}
										position++
										break
									default:
										if buffer[position] != rune('\n') {
											goto l270
										}
										position++
										break
									}
								}

								goto l267
							l270:
								position, tokenIndex, depth = position270, tokenIndex270, depth270
							}
							if !matchDot() {
								goto l267
							}
							goto l266
						l267:
							position, tokenIndex, depth = position267, tokenIndex267, depth267
						}
						goto l263
					l265:
						position, tokenIndex, depth = position263, tokenIndex263, depth263
						{
							position272 := position
							depth++
							if buffer[position] != rune('"') {
								goto l234
							}
							position++
							{
								position275, tokenIndex275, depth275 := position, tokenIndex, depth
								if !rules[ruleEscapedChar]() {
									goto l276
								}
								goto l275
							l276:
								position, tokenIndex, depth = position275, tokenIndex275, depth275
								{
									position279, tokenIndex279, depth279 := position, tokenIndex, depth
									{
										switch buffer[position] {
										case '"':
											if buffer[position] != rune('"') {
												goto l279
											}
											position++
											break
										case '\\':
											if buffer[position] != rune('\\') {
												goto l279
											}
											position++
											break
										default:
											if buffer[position] != rune('\n') {
												goto l279
											}
											position++
											break
										}
									}

									goto l234
								l279:
									position, tokenIndex, depth = position279, tokenIndex279, depth279
								}
								if !matchDot() {
									goto l234
								}
							l277:
								{
									position278, tokenIndex278, depth278 := position, tokenIndex, depth
									{
										position281, tokenIndex281, depth281 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l281
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l281
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l281
												}
												position++
												break
											}
										}

										goto l278
									l281:
										position, tokenIndex, depth = position281, tokenIndex281, depth281
									}
									if !matchDot() {
										goto l278
									}
									goto l277
								l278:
									position, tokenIndex, depth = position278, tokenIndex278, depth278
								}
							}
						l275:
						l273:
							{
								position274, tokenIndex274, depth274 := position, tokenIndex, depth
								{
									position283, tokenIndex283, depth283 := position, tokenIndex, depth
									if !rules[ruleEscapedChar]() {
										goto l284
									}
									goto l283
								l284:
									position, tokenIndex, depth = position283, tokenIndex283, depth283
									{
										position287, tokenIndex287, depth287 := position, tokenIndex, depth
										{
											switch buffer[position] {
											case '"':
												if buffer[position] != rune('"') {
													goto l287
												}
												position++
												break
											case '\\':
												if buffer[position] != rune('\\') {
													goto l287
												}
												position++
												break
											default:
												if buffer[position] != rune('\n') {
													goto l287
												}
												position++
												break
											}
										}

										goto l274
									l287:
										position, tokenIndex, depth = position287, tokenIndex287, depth287
									}
									if !matchDot() {
										goto l274
									}
								l285:
									{
										position286, tokenIndex286, depth286 := position, tokenIndex, depth
										{
											position289, tokenIndex289, depth289 := position, tokenIndex, depth
											{
												switch buffer[position] {
												case '"':
													if buffer[position] != rune('"') {
														goto l289
													}
													position++
													break
												case '\\':
													if buffer[position] != rune('\\') {
														goto l289
													}
													position++
													break
												default:
													if buffer[position] != rune('\n') {
														goto l289
													}
													position++
													break
												}
											}

											goto l286
										l289:
											position, tokenIndex, depth = position289, tokenIndex289, depth289
										}
										if !matchDot() {
											goto l286
										}
										goto l285
									l286:
										position, tokenIndex, depth = position286, tokenIndex286, depth286
									}
								}
							l283:
								goto l273
							l274:
								position, tokenIndex, depth = position274, tokenIndex274, depth274
							}
							if buffer[position] != rune('"') {
								goto l234
							}
							position++
							depth--
							add(ruleQuotedString, position272)
						}
					}
				l263:
					goto l233
				l234:
					position, tokenIndex, depth = position234, tokenIndex234, depth234
				}
				depth--
				add(ruleUntilLineEnd, position232)
			}
			return true
		l231:
			position, tokenIndex, depth = position231, tokenIndex231, depth231
			return false
		},
		/* 20 LineEnd <- <(WS* LineComment? (NL / !.))> */
		func() bool {
			position291, tokenIndex291, depth291 := position, tokenIndex, depth
			{
				position292 := position
				depth++
			l293:
				{
					position294, tokenIndex294, depth294 := position, tokenIndex, depth
					if !rules[ruleWS]() {
						goto l294
					}
					goto l293
				l294:
					position, tokenIndex, depth = position294, tokenIndex294, depth294
				}
				{
					position295, tokenIndex295, depth295 := position, tokenIndex, depth
					if !rules[ruleLineComment]() {
						goto l295
					}
					goto l296
				l295:
					position, tokenIndex, depth = position295, tokenIndex295, depth295
				}
			l296:
				{
					position297, tokenIndex297, depth297 := position, tokenIndex, depth
					if !rules[ruleNL]() {
						goto l298
					}
					goto l297
				l298:
					position, tokenIndex, depth = position297, tokenIndex297, depth297
					{
						position299, tokenIndex299, depth299 := position, tokenIndex, depth
						if !matchDot() {
							goto l299
						}
						goto l291
					l299:
						position, tokenIndex, depth = position299, tokenIndex299, depth299
					}
				}
			l297:
				depth--
				add(ruleLineEnd, position292)
			}
			return true
		l291:
			position, tokenIndex, depth = position291, tokenIndex291, depth291
			return false
		},
		/* 21 LineComment <- <('#' <(!'\n' .)*> Action30)> */
		func() bool {
			position300, tokenIndex300, depth300 := position, tokenIndex, depth
			{
				position301 := position
				depth++
				if buffer[position] != rune('#') {
					goto l300
				}
				position++
				{
					position302 := position
					depth++
				l303:
					{
						position304, tokenIndex304, depth304 := position, tokenIndex, depth
						{
							position305, tokenIndex305, depth305 := position, tokenIndex, depth
							if buffer[position] != rune('\n') {
								goto l305
							}
							position++
							goto l304
						l305:
							position, tokenIndex, depth = position305, tokenIndex305, depth305
						}
						if !matchDot() {
							goto l304
						}
						goto l303
					l304:
						position, tokenIndex, depth = position304, tokenIndex304, depth304
					}
					depth--
					add(rulePegText, position302)
				}
				{
					add(ruleAction30, position)
				}
				depth--
				add(ruleLineComment, position301)
			}
			return true
		l300:
			position, tokenIndex, depth = position300, tokenIndex300, depth300
			return false
		},
		/* 22 BlankLine <- <(((WS LineEnd) / (LineComment? NL)) Action31)> */
		func() bool {
			position307, tokenIndex307, depth307 := position, tokenIndex, depth
			{
				position308 := position
				depth++
				{
					position309, tokenIndex309, depth309 := position, tokenIndex, depth
					if !rules[ruleWS]() {
						goto l310
					}
					if !rules[ruleLineEnd]() {
						goto l310
					}
					goto l309
				l310:
					position, tokenIndex, depth = position309, tokenIndex309, depth309
					{
						position311, tokenIndex311, depth311 := position, tokenIndex, depth
						if !rules[ruleLineComment]() {
							goto l311
						}
						goto l312
					l311:
						position, tokenIndex, depth = position311, tokenIndex311, depth311
					}
				l312:
					if !rules[ruleNL]() {
						goto l307
					}
				}
			l309:
				{
					add(ruleAction31, position)
				}
				depth--
				add(ruleBlankLine, position308)
			}
			return true
		l307:
			position, tokenIndex, depth = position307, tokenIndex307, depth307
			return false
		},
		/* 23 OS <- <(NL / WS)*> */
		func() bool {
			{
				position315 := position
				depth++
			l316:
				{
					position317, tokenIndex317, depth317 := position, tokenIndex, depth
					{
						position318, tokenIndex318, depth318 := position, tokenIndex, depth
						if !rules[ruleNL]() {
							goto l319
						}
						goto l318
					l319:
						position, tokenIndex, depth = position318, tokenIndex318, depth318
						if !rules[ruleWS]() {
							goto l317
						}
					}
				l318:
					goto l316
				l317:
					position, tokenIndex, depth = position317, tokenIndex317, depth317
				}
				depth--
				add(ruleOS, position315)
			}
			return true
		},
		/* 24 WS <- <(' ' / '\t')> */
		func() bool {
			position320, tokenIndex320, depth320 := position, tokenIndex, depth
			{
				position321 := position
				depth++
				{
					position322, tokenIndex322, depth322 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l323
					}
					position++
					goto l322
				l323:
					position, tokenIndex, depth = position322, tokenIndex322, depth322
					if buffer[position] != rune('\t') {
						goto l320
					}
					position++
				}
			l322:
				depth--
				add(ruleWS, position321)
			}
			return true
		l320:
			position, tokenIndex, depth = position320, tokenIndex320, depth320
			return false
		},
		/* 25 UntilNL <- <(!'\n' .)*> */
		nil,
		/* 26 NL <- <('\n' / '\r' / ('\r' '\n'))> */
		func() bool {
			position325, tokenIndex325, depth325 := position, tokenIndex, depth
			{
				position326 := position
				depth++
				{
					position327, tokenIndex327, depth327 := position, tokenIndex, depth
					if buffer[position] != rune('\n') {
						goto l328
					}
					position++
					goto l327
				l328:
					position, tokenIndex, depth = position327, tokenIndex327, depth327
					if buffer[position] != rune('\r') {
						goto l329
					}
					position++
					goto l327
				l329:
					position, tokenIndex, depth = position327, tokenIndex327, depth327
					if buffer[position] != rune('\r') {
						goto l325
					}
					position++
					if buffer[position] != rune('\n') {
						goto l325
					}
					position++
				}
			l327:
				depth--
				add(ruleNL, position326)
			}
			return true
		l325:
			position, tokenIndex, depth = position325, tokenIndex325, depth325
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
		/* 33 Action4 <- <{ p.beginFeature(trimWS(trimWS(p.buf1)), trimWSML(p.buf2), p.buftags); p.buftags = nil }> */
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
	p.rules = rules
}
