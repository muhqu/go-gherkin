package gherkin

import (
	. "github.com/muhqu/go-gherkin/events"
	. "github.com/muhqu/go-gherkin/nodes"
)

type GherkinDOM interface {
	Feature() FeatureNode
}
type GherkinDOMParser interface {
	GherkinParser
	GherkinDOM
	ParseFeature() (FeatureNode, error)
}

type gherkinDOMParser struct {
	gp             GherkinParser
	processed      bool
	feature        MutableFeatureNode
	scenario       MutableScenarioNode
	background     MutableBackgroundNode
	outline        MutableOutlineNode
	step           MutableStepNode
	pyStringIndent int
	pyString       MutablePyStringNode
	table          MutableTableNode
}

func NewGherkinDOMParser(content string) GherkinDOMParser {
	g := &gherkinDOMParser{gp: NewGherkinParser(content)}
	g.gp.WithEventProcessor(g)
	return g
}

func (g *gherkinDOMParser) WithEventProcessor(ep EventProcessor) {
	g.gp.WithEventProcessor(ep)
}

func (g *gherkinDOMParser) WithLogFn(logFn LogFn) {
	g.gp.WithLogFn(logFn)
}

func (g *gherkinDOMParser) Init() {
	g.gp.Init()
}

func (g *gherkinDOMParser) Parse() error {
	return g.gp.Parse()
}

func (g *gherkinDOMParser) Execute() {
	g.gp.Parse()
}

func ParseGherkinFeature(content string) (FeatureNode, error) {
	return NewGherkinDOMParser(content).ParseFeature()
}

func (g *gherkinDOMParser) Feature() FeatureNode {
	if !g.processed {
		_, err := g.ParseFeature()
		if err != nil {
			return nil
		}
	}
	return g.feature
}

func (g *gherkinDOMParser) ParseFeature() (FeatureNode, error) {
	g.processed = true
	g.gp.Init()
	if err := g.gp.Parse(); err != nil {
		return nil, err
	}
	g.gp.Execute()
	return g.feature, nil
}

func (g *gherkinDOMParser) ProcessEvent(event GherkinEvent) {
	switch e := event.(type) {
	// default:
	// 	fmt.Printf("Unexpected Event %T\n", e)

	case *FeatureEvent:
		g.feature = NewMutableFeatureNode(e.Title, e.Description, e.Tags)

	// case *FeatureEndEvent:
	// 	// do nothing

	case *BackgroundEvent:
		node := NewMutableBackgroundNode(e.Title, e.Tags)
		g.scenario = node
		g.feature.SetBackground(node)

	case *ScenarioEvent:
		node := NewMutableScenarioNode(e.Title, e.Tags)
		g.scenario = node
		g.feature.AddScenario(node)

	case *OutlineEvent:
		node := NewMutableOutlineNode(e.Title, e.Tags)
		g.scenario = node
		g.outline = node
		g.feature.AddScenario(node)

	case *OutlineExamplesEvent:
		g.table = nil

	case *OutlineExamplesEndEvent:
		examples := NewOutlineExamplesNode(g.table)
		g.outline.SetExamples(examples)
		g.table = nil

	case *BackgroundEndEvent, *ScenarioEndEvent, *OutlineEndEvent:
		g.scenario = nil
		g.outline = nil
		g.table = nil
		g.pyString = nil

	case *StepEvent:
		g.step = NewMutableStepNode(e.StepType, e.Text)
		g.scenario.AddStep(g.step)

	case *StepEndEvent:
		if g.pyString != nil {
			g.step.WithPyString(g.pyString)
		} else if g.table != nil {
			g.step.WithTable(g.table)
		}
		g.pyString = nil
		g.table = nil
		g.step = nil

	case *TableEvent:
		g.table = NewMutableTableNode()

	case *TableRowEvent:
		g.table.NewRow()

	case *TableCellEvent:
		g.table.AddCell(e.Content)

	// case *TableEndEvent:
	// 	// do nothing

	case *PyStringEvent:
		g.pyStringIndent = len(e.Intent)
		g.pyString = NewMutablePyStringNode()

	case *PyStringLineEvent:
		indent := g.pyStringIndent
		prefix, suffix := e.Line[:indent], e.Line[indent:]
		line := trimLeadingWS(prefix) + suffix
		g.pyString.AddLine(line)

		// case *PyStringEndEvent:
		// 	// do nothing

		// case *BlankLineEvent:
		// 	// TODO: integrate somehow

		// case *CommentEvent:
		// 	// TODO: integrate somehow

	}
}
