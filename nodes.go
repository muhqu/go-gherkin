package gherkin

import (
	"strings"
)

type NodeType int

const (
	FeatureNodeType NodeType = iota
	BackgroundNodeType
	ScenarioNodeType
	StepNodeType
	OutlineNodeType
	OutlineExamplesNodeType
	PyStringNodeType
	TableNodeType
	BlankLineNodeType
	CommentNodeType
)

func (nt NodeType) String() string {
	switch nt {
	case FeatureNodeType:
		return "Feature"
	case BackgroundNodeType:
		return "Background"
	case ScenarioNodeType:
		return "Scenario"
	case StepNodeType:
		return "Step"
	case OutlineNodeType:
		return "Outline"
	case OutlineExamplesNodeType:
		return "OutlineExamples"
	case PyStringNodeType:
		return "PyString"
	case TableNodeType:
		return "Table"
	case BlankLineNodeType:
		return "BlankLine"
	case CommentNodeType:
		return "Comment"
	}
	return "Unknown"
}

type NodeInterface interface {
	NodeType() NodeType
}
type abstractNode struct {
	nodeType NodeType
}

func (a *abstractNode) NodeType() NodeType {
	return a.nodeType
}

// ----------------------------------------

// Representing all Scenarios, Scenario Outlines as well as the Background.
//
//       @tag1 @tag2       <- Tags()
//       Scenario: Title   <- Title()
//          Given ...      <-\
//           When ...         +- Steps()
//           Then ...      <-/
//
type ScenarioNode interface {
	NodeInterface // NodeType: ScenarioNodeType | BackgroundNodeType | OutlineNodeType
	Title() string
	addStep(step StepNode)
	Steps() []StepNode
	Tags() []string
}
type abstractScenarioNode struct {
	abstractNode

	title string
	steps []StepNode
	tags  []string
}

func (a *abstractScenarioNode) Title() string {
	return a.title
}
func (a *abstractScenarioNode) addStep(step StepNode) {
	a.steps = append(a.steps, step)
}
func (a *abstractScenarioNode) Steps() []StepNode {
	return a.steps
}
func (a *abstractScenarioNode) Tags() []string {
	return a.tags
}

// ----------------------------------------

// Representing the Feature
//
//     @tags
//     Feature: Title
//       Description
//
//       Background: ...
//
//       Scenario:  ...
//
type FeatureNode interface {
	NodeInterface // NodeType: FeatureNodeType
	Title() string
	Description() string
	Background() ScenarioNode
	Scenarios() []ScenarioNode
	Tags() []string
}

func newFeatureNode(title, description string, tags []string) *featureNode {
	n := &featureNode{
		title:       trimWS(title),
		description: trimWSML(description),
		tags:        tags,
	}
	return n
}

type featureNode struct {
	abstractNode

	title       string
	description string
	background  *backgroundNode
	scenarios   []ScenarioNode
	tags        []string
}

func (f *featureNode) Title() string {
	return f.title
}
func (f *featureNode) Description() string {
	return f.description
}
func (f *featureNode) Tags() []string {
	return f.tags
}
func (f *featureNode) Background() ScenarioNode {
	if n := f.background; n != nil {
		return n
	}
	return nil
}
func (f *featureNode) Scenarios() []ScenarioNode {
	return f.scenarios
}

// ----------------------------------------

func newBackgroundNode(title string, tags []string) *backgroundNode {
	n := &backgroundNode{}
	n.nodeType = BackgroundNodeType
	n.title = trimWS(title)
	n.tags = tags
	return n
}

type backgroundNode struct {
	abstractScenarioNode
}

// ----------------------------------------

func newScenarioNode(title string, tags []string) *scenarioNode {
	n := &scenarioNode{}
	n.nodeType = ScenarioNodeType
	n.title = trimWS(title)
	n.tags = tags
	return n
}

type scenarioNode struct {
	abstractScenarioNode
}

// ----------------------------------------

// Representing Steps
//
//          StepType   Text
//           |          |
//         .-+-. .------+--------------------------.
//         Given a file with the following contents:
//         '''                                      <
//         All your base are belong to us           <- Argument
//         '''                                      <
type StepNode interface {
	NodeInterface // NodeType: StepNodeType

	StepType() string // Given, When, Then, And, Or, But
	Text() string
	PyString() PyStringNode
	Table() TableNode
}

func newStepNode(stepType, text string) *stepNode {
	n := &stepNode{}
	n.nodeType = StepNodeType
	n.stepType = stepType
	n.text = trimWS(text)
	return n
}

type stepNode struct {
	abstractNode

	stepType string
	text     string
	pyString *pyStringNode
	table    *tableNode
}

func (s *stepNode) StepType() string {
	return s.stepType
}
func (s *stepNode) Text() string {
	return s.text
}
func (s *stepNode) PyString() PyStringNode {
	if p := s.pyString; p != nil {
		return p
	}
	return nil
}
func (s *stepNode) Table() TableNode {
	if t := s.table; t != nil {
		return t
	}
	return nil
}

// ----------------------------------------

type OutlineNode interface {
	ScenarioNode // NodeType: OutlineNodeType

	Examples() OutlineExamplesNode
}

func newOutlineNode(title string, tags []string) *outlineNode {
	n := &outlineNode{}
	n.nodeType = OutlineNodeType
	n.title = trimWS(title)
	n.tags = tags
	return n
}

type outlineNode struct {
	abstractScenarioNode

	examples *outlineExamplesNode
}

func (o *outlineNode) Examples() OutlineExamplesNode {
	return o.examples
}

// ----------------------------------------

type OutlineExamplesNode interface {
	NodeInterface // NodeType: OutlineExamplesNodeType

	Table() TableNode
}

func newOutlineExamplesNode(table *tableNode) *outlineExamplesNode {
	n := &outlineExamplesNode{}
	n.nodeType = OutlineExamplesNodeType
	n.table = table
	return n
}

type outlineExamplesNode struct {
	abstractNode

	table *tableNode
}

func (o *outlineExamplesNode) Table() TableNode {
	if n := o.table; n != nil {
		return n
	}
	return nil
}

// ----------------------------------------

type PyStringNode interface {
	NodeInterface // NodeType: PyStringNodeType

	Lines() []string
	String() string
}

func newPyStringNode(intent int, lines []string) *pyStringNode {
	n := &pyStringNode{}
	n.nodeType = PyStringNodeType
	n.indent = intent
	n.lines = lines
	return n
}

type pyStringNode struct {
	abstractNode

	indent int
	lines  []string
}

func (p *pyStringNode) Lines() []string {
	return p.lines
}

func (p *pyStringNode) String() string {
	return strings.Join(p.lines, "")
}

// ----------------------------------------

type TableNode interface {
	NodeInterface // NodeType: TableNodeType

	Rows() [][]string
}

type tableNode struct {
	abstractNode

	rows [][]string
}

func newTableNode() *tableNode {
	n := &tableNode{}
	n.nodeType = TableNodeType
	return n
}

func (t *tableNode) Rows() [][]string {
	return t.rows
}

// ----------------------------------------

type BlankLineNode interface {
	NodeInterface // NodeType: BlankLineNodeType
}

type blankLineNode struct {
	abstractNode
}

func newBlankLineNode() *blankLineNode {
	n := &blankLineNode{}
	n.nodeType = BlankLineNodeType
	return n
}

// ----------------------------------------

type CommentNode interface {
	NodeInterface // NodeType: CommentNodeType

	Comment() string
}

type commentNode struct {
	abstractNode

	comment string
}

func newCommentNode(comment string) *commentNode {
	n := &commentNode{}
	n.nodeType = CommentNodeType
	n.comment = comment
	return n
}
