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
	Description() string
	Steps() []StepNode
	Tags() []string
	Comment() CommentNode
	Lines() []NodeInterface // StepNode | BlankLineNode
}

type MutableScenarioNode interface {
	ScenarioNode

	AddStep(step StepNode)
	AddBlankLine(line BlankLineNode)
	SetComment(comment CommentNode)
	SetDescription(description string)
}

type abstractScenarioNode struct {
	abstractNode

	title       string
	description string
	steps       []StepNode
	lines       []NodeInterface
	tags        []string
	comment     CommentNode
}

func (a *abstractScenarioNode) Title() string {
	return a.title
}
func (a *abstractScenarioNode) Description() string {
	return a.description
}
func (a *abstractScenarioNode) Steps() []StepNode {
	return a.steps
}
func (a *abstractScenarioNode) Lines() []NodeInterface {
	return a.lines
}
func (a *abstractScenarioNode) Tags() []string {
	return a.tags
}
func (a *abstractScenarioNode) Comment() CommentNode {
	return a.comment
}
func (a *abstractScenarioNode) SetDescription(description string) {
	a.description = description
}
func (a *abstractScenarioNode) AddStep(step StepNode) {
	a.steps = append(a.steps, step)
	a.lines = append(a.lines, step)
}
func (a *abstractScenarioNode) AddBlankLine(line BlankLineNode) {
	a.lines = append(a.lines, line)
}
func (a *abstractScenarioNode) SetComment(comment CommentNode) {
	a.comment = comment
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
	Comment() CommentNode
}

type MutableFeatureNode interface {
	FeatureNode

	SetBackground(background BackgroundNode)
	AddScenario(scenario ScenarioNode)
	SetComment(comment CommentNode)
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
	comment     CommentNode
}

func (f *featureNode) SetBackground(background BackgroundNode) {
	f.background = background
}

func (f *featureNode) AddScenario(scenario ScenarioNode) {
	f.scenarios = append(f.scenarios, scenario)
}

func (f *featureNode) SetComment(comment CommentNode) {
	f.comment = comment
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
func (f *featureNode) Comment() CommentNode {
	return f.comment
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
	Comment() CommentNode
}
type MutableStepNode interface {
	StepNode

	WithStepType(string) MutableStepNode
	WithText(string) MutableStepNode
	SetStepType(string)
	SetText(string)
	WithPyString(PyStringNode) MutableStepNode
	WithTable(TableNode) MutableStepNode
	WithComment(CommentNode) MutableStepNode
	SetPyString(PyStringNode)
	SetTable(TableNode)
	SetComment(CommentNode)
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
	comment  CommentNode
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
func (s *stepNode) WithComment(comment CommentNode) MutableStepNode {
	s.SetComment(comment)
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
func (s *stepNode) SetComment(comment CommentNode) {
	s.comment = comment
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
func (s *stepNode) Comment() CommentNode {
	if c := s.comment; c != nil {
		return c
	}
	return nil
}

// ----------------------------------------

type OutlineNode interface {
	ScenarioNode // NodeType: OutlineNodeType

	Examples() OutlineExamplesNode
	AllExamples() []OutlineExamplesNode
}
type MutableOutlineNode interface {
	OutlineNode

	// MutableScenarioNode can not be imported here
	AddStep(step StepNode)           // stupid
	AddBlankLine(line BlankLineNode) // stupid

	SetDescription(description string)
	SetExamples(examples OutlineExamplesNode)
	AddExamples(examples OutlineExamplesNode)
	SetComment(comment CommentNode)
}

type OutlineExamplesNodes []OutlineExamplesNode

func (o OutlineExamplesNodes) NodeType() NodeType {
	return OutlineExamplesNodeType
}

func (o OutlineExamplesNodes) Table() TableNode {
	t := &tableNode{}
	t.nodeType = TableNodeType
	for _, example := range o {
		t.rows = append(t.rows, example.Table().Rows()...)
	}
	return t
}

func (o OutlineExamplesNodes) Title() string {
	if len(o) > 0 {
		return o[0].Title()
	}
	return ""
}

func (o OutlineExamplesNodes) Comment() CommentNode {
	if len(o) > 0 {
		return o[0].Comment()
	}
	return nil
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

	examples OutlineExamplesNodes
}

func (o *outlineNode) SetExamples(examples OutlineExamplesNode) {
	o.examples = []OutlineExamplesNode{examples}
}

func (o *outlineNode) AddExamples(examples OutlineExamplesNode) {
	o.examples = append(o.examples, examples)
}

func (o *outlineNode) Examples() OutlineExamplesNode {
	return o.examples
}

func (o *outlineNode) AllExamples() []OutlineExamplesNode {
	return []OutlineExamplesNode(o.examples)
}

// ----------------------------------------

type OutlineExamplesNode interface {
	NodeInterface // NodeType: OutlineExamplesNodeType

	Title() string
	Table() TableNode
	Comment() CommentNode
}

type MutableOutlineExamplesNode interface {
	OutlineExamplesNode

	SetTitle(title string)
	SetTable(table TableNode)
	SetComment(comment CommentNode)
}

func NewMutableOutlineExamplesNode(title string) *outlineExamplesNode {
	n := &outlineExamplesNode{}
	n.nodeType = OutlineExamplesNodeType
	n.title = title
	return n
}

func NewOutlineExamplesNode(table TableNode) *outlineExamplesNode {
	n := &outlineExamplesNode{}
	n.nodeType = OutlineExamplesNodeType
	n.table = table
	return n
}

type outlineExamplesNode struct {
	abstractNode

	title   string
	table   TableNode
	comment CommentNode
}

func (o *outlineExamplesNode) Title() string {
	return o.title
}

func (o *outlineExamplesNode) SetTitle(title string) {
	o.title = title
}

func (o *outlineExamplesNode) Table() TableNode {
	if n := o.table; n != nil {
		return n
	}
	return nil
}

func (o *outlineExamplesNode) SetTable(table TableNode) {
	o.table = table
}

func (o *outlineExamplesNode) Comment() CommentNode {
	return o.comment
}

func (o *outlineExamplesNode) SetComment(comment CommentNode) {
	o.comment = comment
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
	RowComments() []CommentNode
}
type MutableTableNode interface {
	TableNode

	WithRows(rows [][]string) MutableTableNode
	NewRow()
	AddRow(row []string)
	AddCell(cell string)
	SetRowComment(comment CommentNode)
}

type tableNode struct {
	abstractNode

	nextRowIndex int
	comments     []CommentNode
	rows         [][]string
}

func NewMutableTableNode() MutableTableNode {
	n := &tableNode{}
	n.nodeType = TableNodeType
	n.comments = make([]CommentNode, 1)
	return n
}

func (t *tableNode) WithRows(rows [][]string) MutableTableNode {
	t.rows = rows
	t.comments = make([]CommentNode, len(rows)+1)
	t.nextRowIndex = len(rows)
	return t
}
func (t *tableNode) NewRow() {
	t.AddRow([]string{})
}
func (t *tableNode) AddRow(row []string) {
	t.nextRowIndex = t.nextRowIndex + 1
	t.rows = append(t.rows, row)
	t.comments = append(t.comments, nil)
}
func (t *tableNode) AddCell(cell string) {
	i := len(t.rows) - 1
	t.rows[i] = append(t.rows[i], cell)
}
func (t *tableNode) RowComments() []CommentNode {
	return t.comments
}
func (t *tableNode) SetRowComment(comment CommentNode) {
	t.comments[t.nextRowIndex-1] = comment
}

func (t *tableNode) Rows() [][]string {
	return t.rows
}

// ----------------------------------------

type BlankLineNode interface {
	NodeInterface // NodeType: BlankLineNodeType

	Comment() CommentNode
}
type MutableBlankLineNode interface {
	BlankLineNode

	SetComment(comment CommentNode)
}

type blankLineNode struct {
	abstractNode

	comment CommentNode
}

func (b *blankLineNode) Comment() CommentNode {
	return b.comment
}

func (b *blankLineNode) SetComment(comment CommentNode) {
	b.comment = comment
}

func NewBlankLineNode() MutableBlankLineNode {
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

func (c *commentNode) Comment() string {
	return c.comment
}

func NewCommentNode(comment string) *commentNode {
	n := &commentNode{}
	n.nodeType = CommentNodeType
	n.comment = comment
	return n
}
