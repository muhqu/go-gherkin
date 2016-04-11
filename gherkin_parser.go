package gherkin

import (
	"strings"

	"github.com/muhqu/go-gherkin/events"
)

type GherkinParser interface {
	WithLogFn(LogFn)
	WithEventProcessor(EventProcessor)
	Init()
	Parse() error
	Execute()
}

func NewGherkinParser(content string) GherkinParser {
	if !strings.HasSuffix(content, "\n") {
		content = content + "\n"
	}

	return &gherkinPegWrapper{gp: &gherkinPeg{Buffer: content}}
}

type gherkinPegWrapper struct {
	gp *gherkinPeg
}

func (gpw *gherkinPegWrapper) WithLogFn(logFn LogFn) {
	gpw.gp.logFn = logFn
}

func (gpw *gherkinPegWrapper) WithEventProcessor(ep EventProcessor) {
	gpw.gp.eventProcessors = append(gpw.gp.eventProcessors, ep)
}

func (gpw *gherkinPegWrapper) Init() {
	gpw.gp.Init()
}

func (gpw *gherkinPegWrapper) Parse() error {
	return gpw.gp.Parse()
}

func (gpw *gherkinPegWrapper) Execute() {
	gpw.gp.Execute()
}

// ----------------------------------------

type EventProcessor interface {
	ProcessEvent(GherkinEvent)
}

type EventProcessorFn func(GherkinEvent)

func (fn EventProcessorFn) ProcessEvent(e GherkinEvent) {
	fn(e)
}

type GherkinEvent events.Event

// ----------------------------------------

type gherkinPegBase struct {
	logFn           LogFn
	eventProcessors []EventProcessor
}

// ----------------------------------------

type LogFn func(msg string, args ...interface{})

func (gp *gherkinPegBase) log(msg string, args ...interface{}) {
	if gp.logFn != nil {
		gp.logFn(msg, args...)
	}
}

func (gp *gherkinPegBase) emit(e GherkinEvent) {
	for _, ep := range gp.eventProcessors {
		ep.ProcessEvent(e)
	}
}

func (gp *gherkinPegBase) beginFeature(title, description string, tags []string) {
	gp.log("BeginFeature: %#v: %#v tags:%+v", title, description, tags)
	gp.emit(&events.FeatureEvent{title, description, tags})
}
func (gp *gherkinPegBase) endFeature() {
	gp.log("EndFeature")
	gp.emit(&events.FeatureEndEvent{})
}

func (gp *gherkinPegBase) beginBackground(title, description string, tags []string) {
	gp.log("BeginBackground: %#v: %#v tags:%+v", title, description, tags)
	gp.emit(&events.BackgroundEvent{title, description, tags})
}
func (gp *gherkinPegBase) endBackground() {
	gp.log("EndBackground")
	gp.emit(&events.BackgroundEndEvent{})
}

func (gp *gherkinPegBase) beginScenario(title, description string, tags []string) {
	gp.log("BeginScenario: %#v: %#v tags:%+v", title, description, tags)
	gp.emit(&events.ScenarioEvent{title, description, tags})
}
func (gp *gherkinPegBase) endScenario() {
	gp.log("EndScenario")
	gp.emit(&events.ScenarioEndEvent{})
}

func (gp *gherkinPegBase) beginOutline(title, description string, tags []string) {
	gp.log("BeginOutline: %#v: %#v tags:%+v", title, description, tags)
	gp.emit(&events.OutlineEvent{title, description, tags})
}
func (gp *gherkinPegBase) endOutline() {
	gp.log("EndOutline")
	gp.emit(&events.OutlineEndEvent{})
}

func (gp *gherkinPegBase) beginOutlineExamples(title string) {
	gp.log("BeginOutlineExamples")
	gp.emit(&events.OutlineExamplesEvent{title})
}
func (gp *gherkinPegBase) endOutlineExamples() {
	gp.log("EndOutlineExamples")
	gp.emit(&events.OutlineExamplesEndEvent{})
}

func (gp *gherkinPegBase) beginStep(stepType, name string) {
	gp.log("BeginStep: %#v: %#v", stepType, name)
	gp.emit(&events.StepEvent{stepType, name})
}
func (gp *gherkinPegBase) endStep() {
	gp.log("EndStep")
	gp.emit(&events.StepEndEvent{})
}

func (gp *gherkinPegBase) beginPyString(indent string) {
	width := len(trimNL(indent))
	gp.log("BeginPyString: indent=%d", width)
	gp.emit(&events.PyStringEvent{indent})
}
func (gp *gherkinPegBase) bufferPyString(line string) {
	gp.log("BufferPyString: %#v", line)
	gp.emit(&events.PyStringLineEvent{line})
	/*
		indent := gp.pyString.indent
		prefix, suffix := line[:indent], line[indent:]
		newline := trimLeadingWS(prefix) + suffix
		//gp.log("BufferPyString: PREFIX: '%#v' SUFFIX: '%#v' LINE: '%#v'", prefix, suffix, newline)
		gp.pyString.lines = append(gp.pyString.lines, newline)
	*/
}
func (gp *gherkinPegBase) endPyString() {
	gp.log("EndPyString")
	gp.emit(&events.PyStringEndEvent{})
}

func (gp *gherkinPegBase) beginTable() {
	gp.log("BeginTable")
	gp.emit(&events.TableEvent{})
}
func (gp *gherkinPegBase) beginTableRow() {
	gp.log("BeginTableRow")
	gp.emit(&events.TableRowEvent{})
}
func (gp *gherkinPegBase) beginTableCell() {
	gp.log("BeginTableCell")
}
func (gp *gherkinPegBase) endTableCell(buf string) {
	gp.log("EndTableCell: %#v", buf)
	gp.emit(&events.TableCellEvent{buf})
}
func (gp *gherkinPegBase) endTableRow() {
	gp.log("EndTableRow")
	gp.emit(&events.TableRowEndEvent{})
}
func (gp *gherkinPegBase) endTable() {
	gp.log("EndTable")
	gp.emit(&events.TableEndEvent{})
}

func (gp *gherkinPegBase) triggerComment(comment string) {
	gp.log("triggerComment")
	gp.emit(&events.CommentEvent{comment})
}
func (gp *gherkinPegBase) triggerBlankLine() {
	gp.log("triggerBlankLine")
	gp.emit(&events.BlankLineEvent{})
}
