// Sub-Package gherkin/nodes provides the data-structure types for the gherkin DOM parser.
package nodes

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
	Steps() []StepNode
	Tags() []string
}

type MutableScenarioNode interface {
	ScenarioNode

	AddStep(step StepNode)
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
func (a *abstractScenarioNode) Steps() []StepNode {
	return a.steps
}
func (a *abstractScenarioNode) Tags() []string {
	return a.tags
}
func (a *abstractScenarioNode) AddStep(step StepNode) {
	a.steps = append(a.steps, step)
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
	Background() BackgroundNode
	Scenarios() []ScenarioNode
	Tags() []string
}

type MutableFeatureNode interface {
	FeatureNode

	SetBackground(background BackgroundNode)
	AddScenario(scenario ScenarioNode)
}

func NewMutableFeatureNode(title, description string, tags []string) MutableFeatureNode {
	n := &featureNode{
		title:       title,
		description: description,
		tags:        tags,
	}
	return n
}

type featureNode struct {
	abstractNode

	title       string
	description string
	background  BackgroundNode
	scenarios   []ScenarioNode
	tags        []string
}

func (f *featureNode) SetBackground(background BackgroundNode) {
	f.background = background
}

func (f *featureNode) AddScenario(scenario ScenarioNode) {
	f.scenarios = append(f.scenarios, scenario)
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
func (f *featureNode) Background() BackgroundNode {
	if n := f.background; n != nil {
		return n
	}
	return nil
}
func (f *featureNode) Scenarios() []ScenarioNode {
	return f.scenarios
}

// ----------------------------------------

type BackgroundNode interface {
	ScenarioNode
}

type MutableBackgroundNode interface {
	MutableScenarioNode
}

func NewMutableBackgroundNode(title string, tags []string) MutableBackgroundNode {
	n := &backgroundNode{}
	n.nodeType = BackgroundNodeType
	n.title = title
	n.tags = tags
	return n
}

type backgroundNode struct {
	abstractScenarioNode
}

// ----------------------------------------

func NewMutableScenarioNode(title string, tags []string) MutableScenarioNode {
	n := &scenarioNode{}
	n.nodeType = ScenarioNodeType
	n.title = title
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
type MutableStepNode interface {
	StepNode

	WithStepType(string) MutableStepNode
	WithText(string) MutableStepNode
	SetStepType(string)
	SetText(string)
	WithPyString(PyStringNode) MutableStepNode
	WithTable(TableNode) MutableStepNode
	SetPyString(PyStringNode)
	SetTable(TableNode)
}

func NewMutableStepNode(stepType, text string) MutableStepNode {
	n := &stepNode{}
	n.nodeType = StepNodeType
	n.SetStepType(stepType)
	n.SetText(text)
	return n
}

type stepNode struct {
	abstractNode

	stepType string
	text     string
	pyString PyStringNode
	table    TableNode
}

func (s *stepNode) SetStepType(stepType string) {
	s.stepType = stepType
}
func (s *stepNode) SetText(text string) {
	s.text = text
}
func (s *stepNode) WithStepType(stepType string) MutableStepNode {
	s.SetStepType(stepType)
	return s
}
func (s *stepNode) WithText(text string) MutableStepNode {
	s.SetText(text)
	return s
}
func (s *stepNode) WithPyString(pyString PyStringNode) MutableStepNode {
	s.SetPyString(pyString)
	return s
}
func (s *stepNode) WithTable(table TableNode) MutableStepNode {
	s.SetTable(table)
	return s
}
func (s *stepNode) SetPyString(pyString PyStringNode) {
	s.pyString = pyString
	s.table = nil
}
func (s *stepNode) SetTable(table TableNode) {
	s.table = table
	s.pyString = nil
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
type MutableOutlineNode interface {
	OutlineNode
	// MutableScenarioNode
	AddStep(step StepNode) // stupid
	SetExamples(examples OutlineExamplesNode)
}

func NewMutableOutlineNode(title string, tags []string) MutableOutlineNode {
	n := &outlineNode{}
	n.nodeType = OutlineNodeType
	n.title = title
	n.tags = tags
	return n
}

type outlineNode struct {
	abstractScenarioNode

	examples OutlineExamplesNode
}

func (o *outlineNode) SetExamples(examples OutlineExamplesNode) {
	o.examples = examples
}

func (o *outlineNode) Examples() OutlineExamplesNode {
	return o.examples
}

// ----------------------------------------

type OutlineExamplesNode interface {
	NodeInterface // NodeType: OutlineExamplesNodeType

	Table() TableNode
}

func NewOutlineExamplesNode(table TableNode) *outlineExamplesNode {
	n := &outlineExamplesNode{}
	n.nodeType = OutlineExamplesNodeType
	n.table = table
	return n
}

type outlineExamplesNode struct {
	abstractNode

	table TableNode
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

type MutablePyStringNode interface {
	PyStringNode

	AddLine(line string)
	WithLines(lines []string) MutablePyStringNode
}

func NewMutablePyStringNode() MutablePyStringNode {
	n := &pyStringNode{}
	n.nodeType = PyStringNodeType
	return n
}

type pyStringNode struct {
	abstractNode

	lines []string
}

func (p *pyStringNode) AddLine(line string) {
	p.lines = append(p.lines, line)
}
func (p *pyStringNode) WithLines(lines []string) MutablePyStringNode {
	p.lines = lines
	return p
}

func (p *pyStringNode) Lines() []string {
	return p.lines
}

func (p *pyStringNode) String() string {
	s := ""
	for _, line := range p.lines {
		s += line + "\n"
	}
	return s
}

// ----------------------------------------

type TableNode interface {
	NodeInterface // NodeType: TableNodeType

	Rows() [][]string
}
type MutableTableNode interface {
	TableNode

	WithRows(rows [][]string) MutableTableNode
	NewRow()
	AddRow(row []string)
	AddCell(cell string)
}

type tableNode struct {
	abstractNode

	rows [][]string
}

func NewMutableTableNode() MutableTableNode {
	n := &tableNode{}
	n.nodeType = TableNodeType
	return n
}

func (t *tableNode) WithRows(rows [][]string) MutableTableNode {
	t.rows = rows
	return t
}
func (t *tableNode) NewRow() {
	t.AddRow([]string{})
}
func (t *tableNode) AddRow(row []string) {
	t.rows = append(t.rows, row)
}
func (t *tableNode) AddCell(cell string) {
	i := len(t.rows) - 1
	t.rows[i] = append(t.rows[i], cell)
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

func NewBlankLineNode() *blankLineNode {
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

func NewCommentNode(comment string) *commentNode {
	n := &commentNode{}
	n.nodeType = CommentNodeType
	n.comment = comment
	return n
}
