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
	f := newFeatureNode(title, description, tags)
	gp.log("BeginFeature: %#v: %#v tags:%+v", f.Title(), f.Description(), f.Tags())
	gp.feature = f
	gp.emit(BeginNodeEvent(gp.feature))
}
func (gp *gherkinPegBase) endFeature() {
	gp.log("EndFeature")
	gp.emit(EndNodeEvent(gp.feature))
}

// func (gp *gherkinPegBase) addScenario(scenarioNode ScenarioNode) {
// 	gp.feature.scenarios = append(gp.feature.scenarios, scenarioNode)
// 	gp.scenario = scenarioNode
// }

func (gp *gherkinPegBase) beginBackground(title string, tags []string) {
	gp.log("BeginBackground: %#v tags:%+v", title, tags)
	scenario := newBackgroundNode(title, tags)
	gp.scenario = scenario
	//gp.feature.background = scenario
	gp.emit(BeginNodeEvent(gp.scenario))
}
func (gp *gherkinPegBase) endBackground() {
	gp.log("EndBackground")
	gp.emit(EndNodeEvent(gp.scenario))
}

func (gp *gherkinPegBase) beginScenario(title string, tags []string) {
	gp.log("BeginScenario: %#v tags:%+v", title, tags)
	gp.scenario = newScenarioNode(title, tags)
	//gp.addScenario(scenario)
	gp.emit(BeginNodeEvent(gp.scenario))
}
func (gp *gherkinPegBase) endScenario() {
	gp.log("EndScenario")
	gp.emit(EndNodeEvent(gp.scenario))
}

func (gp *gherkinPegBase) beginOutline(title string, tags []string) {
	gp.log("BeginOutline: %#v tags:%+v", title, tags)
	scenario := newOutlineNode(title, tags)
	//gp.addScenario(scenario)
	gp.outline = scenario
	gp.emit(BeginNodeEvent(gp.outline))
}
func (gp *gherkinPegBase) endOutline() {
	gp.log("EndOutline")
	gp.emit(EndNodeEvent(gp.outline))
	gp.outline = nil
}

func (gp *gherkinPegBase) beginOutlineExamples() {
	gp.log("BeginOutlineExamples")
	gp.table = nil
}
func (gp *gherkinPegBase) endOutlineExamples() {
	gp.log("EndOutlineExamples")
	gp.outline.examples = newOutlineExamplesNode(gp.table)
	gp.table = nil
}

func (gp *gherkinPegBase) beginStep(stepType, name string) {
	gp.log("BeginStep: %#v: %#v", stepType, name)
	step := newStepNode(stepType, name)
	gp.step = step
	//gp.scenario.addStep(step)
	gp.emit(BeginNodeEvent(gp.step))
}
func (gp *gherkinPegBase) endStep() {
	gp.log("EndStep")
	if gp.pyString != nil {
		gp.step.pyString = gp.pyString
		gp.pyString = nil
	}
	if gp.table != nil {
		gp.step.table = gp.table
		gp.table = nil
	}
	gp.emit(EndNodeEvent(gp.step))
	gp.step = nil
}

func (gp *gherkinPegBase) beginPyString(indent string) {
	width := len(trimNL(indent))
	gp.log("BeginPyString: indent=%d", width)
	pyString := newPyStringNode(width, nil)
	gp.pyString = pyString
	gp.emit(BeginNodeEvent(gp.pyString))
}
func (gp *gherkinPegBase) bufferPyString(line string) {
	gp.log("BufferPyString: %#v", line)
	indent := gp.pyString.indent
	prefix, suffix := line[:indent], line[indent:]
	newline := trimLeadingWS(prefix) + suffix
	//gp.log("BufferPyString: PREFIX: '%#v' SUFFIX: '%#v' LINE: '%#v'", prefix, suffix, newline)
	gp.pyString.lines = append(gp.pyString.lines, newline)
}
func (gp *gherkinPegBase) endPyString() {
	gp.log("EndPyString")
	gp.emit(EndNodeEvent(gp.pyString))
}

func (gp *gherkinPegBase) beginTable() {
	gp.log("BeginTable")
	table := newTableNode()
	gp.table = table
	gp.emit(BeginNodeEvent(gp.table))
}
func (gp *gherkinPegBase) beginTableRow() {
	gp.log("BeginTableRow")
	gp.table.rows = append(gp.table.rows, []string{})
}
func (gp *gherkinPegBase) beginTableCell() {
	gp.log("BeginTableCell")
}
func (gp *gherkinPegBase) endTableCell(buf string) {
	gp.log("EndTableCell: %#v", buf)
	rows := gp.table.rows
	i := len(rows) - 1
	rows[i] = append(rows[i], trimWS(buf))
}
func (gp *gherkinPegBase) endTableRow() {
	gp.log("EndTableRow")
}
func (gp *gherkinPegBase) endTable() {
	gp.log("EndTable")
	gp.emit(EndNodeEvent(gp.table))
}
