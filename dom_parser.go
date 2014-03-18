package gherkin

import (
	"io"
)

type GherkinDOM interface {
	Feature() FeatureNode
	Format(f GherkinFormater) io.Reader
}
type GherkinDOMParser interface {
	GherkinParser
	GherkinDOM
	ParseFeature() (FeatureNode, error)
}

type gherkinDOMParser struct {
	gp        GherkinParser
	processed bool
	feature   *featureNode
	scenario  ScenarioNode
	outline   *outlineNode
	step      *stepNode
	pyString  *pyStringNode
	table     *tableNode
}

func NewGherkinDOMParser(content string) GherkinDOMParser {
	g := &gherkinDOMParser{gp: NewGherkinParser(content)}
	g.gp.WithNodeEventProcessor(g)
	return g
}

func (g *gherkinDOMParser) WithNodeEventProcessor(ep NodeEventProcessor) {
	g.gp.WithNodeEventProcessor(ep)
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

func (g *gherkinDOMParser) ProcessNodeEvent(ne NodeEvent) {
	switch e := ne.(type) {
	// default:
	// 	fmt.Printf("Unexpected Event %T\n", e)

	case *FeatureEvent:
		g.feature = newFeatureNode(e.Title, e.Description, e.Tags)

	// case *FeatureEndEvent:
	// 	// do nothing

	case *BackgroundEvent:
		node := newBackgroundNode(e.Title, e.Tags)
		g.scenario = node
		g.feature.background = node

	case *ScenarioEvent:
		node := newScenarioNode(e.Title, e.Tags)
		g.scenario = node
		g.feature.scenarios = append(g.feature.scenarios, node)

	case *OutlineEvent:
		node := newOutlineNode(e.Title, e.Tags)
		g.scenario = node
		g.outline = node
		g.feature.scenarios = append(g.feature.scenarios, node)

	case *OutlineExamplesEvent:
		g.table = nil

	case *OutlineExamplesEndEvent:
		g.outline.examples = newOutlineExamplesNode(g.table)
		g.table = nil

	case *BackgroundEndEvent, *ScenarioEndEvent, *OutlineEndEvent:
		g.scenario = nil
		g.outline = nil
		g.table = nil
		g.pyString = nil

	case *StepEvent:
		node := newStepNode(e.StepType, e.Text)
		g.step = node
		g.scenario.addStep(g.step)

	case *StepEndEvent:
		if g.pyString != nil {
			g.step.pyString = g.pyString
		} else if g.table != nil {
			g.step.table = g.table
		}
		g.pyString = nil
		g.table = nil
		g.step = nil

	case *TableEvent:
		g.table = newTableNode()

	case *TableRowEvent:
		g.table.rows = append(g.table.rows, []string{})

	case *TableCellEvent:
		rows := g.table.rows
		i := len(rows) - 1
		rows[i] = append(rows[i], e.Content)

	// case *TableEndEvent:
	// 	// do nothing

	case *PyStringEvent:
		node := newPyStringNode(len(e.Intent), []string{})
		g.pyString = node

	case *PyStringLineEvent:
		indent := g.pyString.indent
		prefix, suffix := e.Line[:indent], e.Line[indent:]
		line := trimLeadingWS(prefix) + suffix
		g.pyString.lines = append(g.pyString.lines, line)

		// case *PyStringEndEvent:
		// 	// do nothing

		// case *BlankLineEvent:
		// 	// TODO: integrate somehow

		// case *CommentEvent:
		// 	// TODO: integrate somehow

	}
}
