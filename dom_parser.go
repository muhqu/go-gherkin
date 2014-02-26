package gherkin

import (
	"fmt"
)

type GherkinDOM interface {
	Feature() FeatureNode
}
type GherkinDOMParser interface {
	GherkinParser
	GherkinDOM
	GherkinDOMWriter
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

func (g *gherkinDOMParser) ProcessNodeEvent(e NodeEvent) {
	et := e.EventType()
	switch node := e.Node().(type) {
	default:
		fmt.Printf("unexpected type %T", node)
	case *featureNode:
		switch et {
		case BeginNodeEventType:
			g.feature = node
		}
	case *backgroundNode:
		switch et {
		case BeginNodeEventType:
			g.scenario = node
			g.feature.background = node
		}
	case *scenarioNode:
		switch et {
		case BeginNodeEventType:
			g.scenario = node
			g.feature.scenarios = append(g.feature.scenarios, node)
		}
	case *outlineNode:
		switch et {
		case BeginNodeEventType:
			g.scenario = node
			g.feature.scenarios = append(g.feature.scenarios, node)
		}
	case *stepNode:
		switch et {
		case BeginNodeEventType:
			g.step = node
			g.scenario.addStep(g.step)
		case EndNodeEventType:
			if g.pyString != nil {
				g.step.pyString = g.pyString
				g.pyString = nil
			} else if g.table != nil {
				g.step.table = g.table
				g.table = nil
			}
		}
	case *tableNode:
		g.table = node
	case *pyStringNode:
		g.pyString = node
	}
}
