package gherkin

type GherkinParser interface {
	WithLogFn(LogFn)
	WithNodeEventProcessor(NodeEventProcessor)
	Init()
	Parse() error
	Execute()
}

func NewGherkinParser(content string) GherkinParser {
	return &gherkinPegWrapper{gp: &gherkinPeg{Buffer: content}}
}

type gherkinPegWrapper struct {
	gp *gherkinPeg
}

func (gpw *gherkinPegWrapper) WithLogFn(logFn LogFn) {
	gpw.gp.logFn = logFn
}

func (gpw *gherkinPegWrapper) WithNodeEventProcessor(ep NodeEventProcessor) {
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

func (gpw *gherkinPegWrapper) Feature() FeatureNode {
	return gpw.gp.feature
}

// ----------------------------------------

type gherkinPegBase struct {
	logFn           LogFn
	eventProcessors []NodeEventProcessor

	feature  *featureNode
	scenario ScenarioNode
	outline  *outlineNode
	step     *stepNode
	pyString *pyStringNode
	table    *tableNode
}

// ----------------------------------------

type LogFn func(msg string, args ...interface{})

func (gp *gherkinPegBase) log(msg string, args ...interface{}) {
	if gp.logFn != nil {
		gp.logFn(msg, args...)
	}
}

func (gp *gherkinPegBase) emit(e NodeEvent) {
	for _, ep := range gp.eventProcessors {
		ep.ProcessNodeEvent(e)
	}
}

func (gp *gherkinPegBase) beginFeature(title string, description string, tags []string) {
	gp.log("BeginFeature: %#v: %#v tags:%+v", title, description, tags)
	gp.emit(&FeatureEvent{title, description, tags})
}
func (gp *gherkinPegBase) endFeature() {
	gp.log("EndFeature")
	gp.emit(&FeatureEndEvent{})
}

func (gp *gherkinPegBase) beginBackground(title string, tags []string) {
	gp.log("BeginBackground: %#v tags:%+v", title, tags)
	gp.emit(&BackgroundEvent{title, tags})
}
func (gp *gherkinPegBase) endBackground() {
	gp.log("EndBackground")
	gp.emit(&BackgroundEndEvent{})
}

func (gp *gherkinPegBase) beginScenario(title string, tags []string) {
	gp.log("BeginScenario: %#v tags:%+v", title, tags)
	gp.emit(&ScenarioEvent{title, tags})
}
func (gp *gherkinPegBase) endScenario() {
	gp.log("EndScenario")
	gp.emit(&ScenarioEndEvent{})
}

func (gp *gherkinPegBase) beginOutline(title string, tags []string) {
	gp.log("BeginOutline: %#v tags:%+v", title, tags)
	gp.emit(&OutlineEvent{title, tags})
}
func (gp *gherkinPegBase) endOutline() {
	gp.log("EndOutline")
	gp.emit(&OutlineEndEvent{})
}

func (gp *gherkinPegBase) beginOutlineExamples() {
	gp.log("BeginOutlineExamples")
	gp.emit(&OutlineExamplesEvent{})
}
func (gp *gherkinPegBase) endOutlineExamples() {
	gp.log("EndOutlineExamples")
	gp.emit(&OutlineExamplesEndEvent{})
}

func (gp *gherkinPegBase) beginStep(stepType, name string) {
	gp.log("BeginStep: %#v: %#v", stepType, name)
	gp.emit(&StepEvent{stepType, name})
}
func (gp *gherkinPegBase) endStep() {
	gp.log("EndStep")
	gp.emit(&StepEndEvent{})
}

func (gp *gherkinPegBase) beginPyString(indent string) {
	width := len(trimNL(indent))
	gp.log("BeginPyString: indent=%d", width)
	gp.emit(&PyStringEvent{indent})
}
func (gp *gherkinPegBase) bufferPyString(line string) {
	gp.log("BufferPyString: %#v", line)
	gp.emit(&PyStringLineEvent{line})
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
	gp.emit(&PyStringEndEvent{})
}

func (gp *gherkinPegBase) beginTable() {
	gp.log("BeginTable")
	gp.emit(&TableEvent{})
}
func (gp *gherkinPegBase) beginTableRow() {
	gp.log("BeginTableRow")
	gp.emit(&TableRowEvent{})
}
func (gp *gherkinPegBase) beginTableCell() {
	gp.log("BeginTableCell")
}
func (gp *gherkinPegBase) endTableCell(buf string) {
	gp.log("EndTableCell: %#v", buf)
	gp.emit(&TableCellEvent{buf})
	/*
		rows := gp.table.rows
		i := len(rows) - 1
		rows[i] = append(rows[i], trimWS(buf))
	*/
}
func (gp *gherkinPegBase) endTableRow() {
	gp.log("EndTableRow")
}
func (gp *gherkinPegBase) endTable() {
	gp.log("EndTable")
	gp.emit(&TableEndEvent{})
}

func (gp *gherkinPegBase) triggerComment(comment string) {
	gp.log("triggerComment")
	gp.emit(&CommentEvent{comment})
}
func (gp *gherkinPegBase) triggerBlankLine() {
	gp.log("triggerBlankLine")
	gp.emit(&BlankLineEvent{})
}
