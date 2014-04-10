// Sub-Package gherkin/formater provides everything to pretty-print a Gherkin DOM.
package formater

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/muhqu/go-gherkin"
	"github.com/muhqu/go-gherkin/nodes"
)

type GherkinFormater interface {
	Format(gherkin.GherkinDOM, io.Writer)
	FormatFeature(nodes.FeatureNode, io.Writer)
	FormatScenario(nodes.ScenarioNode, io.Writer)
	FormatStep(nodes.StepNode, io.Writer)
	FormatTable(nodes.TableNode, io.Writer)
	FormatPyString(nodes.PyStringNode, io.Writer)
}

// -------------------

type GherkinPrettyFormater struct {
	AnsiColors             bool
	CenterSteps            bool
	SkipSteps              bool
	SkipComments           bool
	NoAlignComments        bool
	AlignCommentsMinIndent int
}

const AlignCommentsMinIndentDefault = 45

type gherkinPrettyPrinter struct {
	gpf *GherkinPrettyFormater
	io.Writer

	linebuff lineBuffer
}

type lineBuffer interface {
	Writeln(str fmt.Stringer, comment ...fmt.Stringer)
	Flush()
}
type strLenCmt struct {
	str     string
	width   int
	comment string
}
type lineCommentAlignmentBuffer struct {
	io.Writer

	minwidth int
	maxwidth int
	lines    []*strLenCmt
}

func (l *lineCommentAlignmentBuffer) Writeln(stringer fmt.Stringer, comments ...fmt.Stringer) {
	str := stringer.String()
	var width int
	if t, ok := stringer.(*styledString); ok {
		width = t.width
	} else {
		width = len(str)
	}
	comment := ""
	for _, value := range comments {
		comment += value.String()
	}
	if l.maxwidth < width {
		l.maxwidth = width
	}
	l.lines = append(l.lines, &strLenCmt{str, width, comment})
	if comment == "" {
		l.Flush()
	}
}
func (l *lineCommentAlignmentBuffer) Flush() {
	if l.maxwidth > 0 && l.maxwidth < l.minwidth {
		l.maxwidth = l.minwidth
	}
	for _, line := range l.lines {
		if line.comment == "" {
			fmt.Fprintf(l.Writer, "%s\n", line.str)
		} else {
			fmtStr := fmt.Sprintf("%%s%%-%ds%%s\n", l.maxwidth-line.width+1)
			if line.width == 0 {
				fmtStr = "%s%s%s\n"
			}
			fmt.Fprintf(l.Writer, fmtStr, line.str, "", line.comment)
		}
	}
	l.lines = nil
	l.maxwidth = 0
}

type noCommentLineBuffer struct {
	io.Writer
}

func (l *noCommentLineBuffer) Writeln(stringer fmt.Stringer, comments ...fmt.Stringer) {
	str := stringer.String()
	fmt.Fprintf(l.Writer, "%s\n", str)
}
func (l *noCommentLineBuffer) Flush() {
}

type noAlignCommentLineBuffer struct {
	io.Writer
}

func (l *noAlignCommentLineBuffer) Writeln(stringer fmt.Stringer, comments ...fmt.Stringer) {
	str := stringer.String()
	comment := ""
	for _, value := range comments {
		comment += value.String()
	}
	if comment == "" {
		fmt.Fprintf(l.Writer, "%s\n", str)
	} else {
		fmt.Fprintf(l.Writer, "%s #%s\n", str, comment)
	}

}
func (l *noAlignCommentLineBuffer) Flush() {
}

func newGherkinPrettyPrinter(gpf *GherkinPrettyFormater, out io.Writer) *gherkinPrettyPrinter {
	g := &gherkinPrettyPrinter{}
	g.gpf = gpf
	g.Writer = out
	if gpf.SkipComments {
		g.linebuff = &noCommentLineBuffer{out}
	} else if gpf.NoAlignComments {
		g.linebuff = &noAlignCommentLineBuffer{out}
	} else {
		l := &lineCommentAlignmentBuffer{}
		l.Writer = out
		if gpf.AlignCommentsMinIndent > 0 {
			l.minwidth = gpf.AlignCommentsMinIndent
		} else {
			l.minwidth = AlignCommentsMinIndentDefault
		}
		g.linebuff = l
	}
	return g
}

func (gpf *GherkinPrettyFormater) Format(gd gherkin.GherkinDOM, out io.Writer) {
	gpf.FormatFeature(gd.Feature(), out)
}

func (gpf *GherkinPrettyFormater) FormatFeature(node nodes.FeatureNode, out io.Writer) {
	g := newGherkinPrettyPrinter(gpf, out)
	g.FormatFeature(node)
}
func (gpf *GherkinPrettyFormater) FormatScenario(node nodes.ScenarioNode, out io.Writer) {
	g := newGherkinPrettyPrinter(gpf, out)
	g.FormatScenario(node)
}
func (gpf *GherkinPrettyFormater) FormatStep(node nodes.StepNode, out io.Writer) {
	g := newGherkinPrettyPrinter(gpf, out)
	g.FormatStep(node)
}
func (gpf *GherkinPrettyFormater) FormatTable(node nodes.TableNode, out io.Writer) {
	g := newGherkinPrettyPrinter(gpf, out)
	g.FormatTable(node)
}
func (gpf *GherkinPrettyFormater) FormatPyString(node nodes.PyStringNode, out io.Writer) {
	g := newGherkinPrettyPrinter(gpf, out)
	g.FormatPyString(node)
}

func (g *gherkinPrettyPrinter) write(s string) {
	io.WriteString(g.Writer, s)
}

type ansiStyle string

const (
	c_BOLD         ansiStyle = "1"
	c_WHITE                  = "29"
	c_BLACK                  = "30"
	c_RED                    = "31"
	c_GREEN                  = "32"
	c_YELLOW                 = "33"
	c_BLUE                   = "34"
	c_MAGENTA                = "35"
	c_CYAN                   = "36"
	c_GRAY                   = "30;1"
	c_BOLD_RED               = "31;1"
	c_BOLD_GREEN             = "32;1"
	c_BOLD_YELLOW            = "33;1"
	c_BOLD_BLUE              = "34;1"
	c_BOLD_MAGENTA           = "35;1"
	c_BOLD_CYAN              = "36;1"
)

type styledString struct {
	str   string
	width int
}

func (s *styledString) String() string {
	return s.str
}

func (g *gherkinPrettyPrinter) joinStyledStrings(args ...*styledString) *styledString {
	s := &styledString{}
	for _, arg := range args {
		s.str += arg.str
		s.width += arg.width
	}
	return s
}

func (g *gherkinPrettyPrinter) colored(color ansiStyle, str string, args ...interface{}) *styledString {
	var formatedStr string
	if args != nil {
		str = fmt.Sprintf(str, args...)
	}
	if g.gpf.AnsiColors {
		formatedStr = fmt.Sprintf("\x1B[%sm%s\x1B[m", color, str)
	} else {
		formatedStr = str
	}
	return &styledString{formatedStr, utf8.RuneCountInString(str)}
}

func (g *gherkinPrettyPrinter) FormatFeature(node nodes.FeatureNode) {
	tags := node.Tags()
	if len(tags) > 0 {
		g.write(fmt.Sprintf("%s\n", g.colored(c_CYAN, fmtTags(tags))))
	}
	g.linebuff.Writeln(g.joinStyledStrings(
		g.colored(c_BOLD, "%s", "Feature:"),
		g.colored(c_WHITE, " %s", node.Title()),
	),
		g.coloredComment(node.Comment()),
	)
	g.linebuff.Flush()
	if node.Description() != "" {
		g.write(prefixLines("  ", node.Description()) + "\n")
	}
	g.write("\n")

	if !g.gpf.SkipSteps && node.Background() != nil {
		g.FormatScenario(node.Background())
		g.write("\n")
	}

	for _, scenario := range node.Scenarios() {
		g.FormatScenario(scenario)
		g.write("\n")
	}
}

func (g *gherkinPrettyPrinter) FormatScenario(node nodes.ScenarioNode) {
	tags := node.Tags()
	if len(tags) > 0 {
		g.write(fmt.Sprintf("  %s\n", g.colored(c_CYAN, fmtTags(tags))))
	}
	switch node.NodeType() {
	case nodes.BackgroundNodeType:
		g.linebuff.Writeln(g.colored(c_BOLD, "  %s", "Background:"), g.coloredComment(node.Comment()))
	case nodes.ScenarioNodeType:
		g.linebuff.Writeln(g.joinStyledStrings(
			g.colored(c_BOLD, "  %s", "Scenario:"),
			g.colored(c_WHITE, " %s", node.Title()),
		),
			g.coloredComment(node.Comment()),
		)
	case nodes.OutlineNodeType:
		g.write(fmt.Sprintf("  %s %s\n", g.colored(c_BOLD, "Scenario Outline:"), g.colored(c_WHITE, node.Title())))
	}
	if !g.gpf.SkipSteps {
		for _, line := range node.Lines() {
			if step, ok := line.(nodes.StepNode); ok {
				g.formatStep(step)
			} else if blankLine, ok := line.(nodes.BlankLineNode); ok {
				g.linebuff.Writeln(&styledString{"  ", 0}, g.coloredComment(blankLine.Comment()))
			}
		}
		if node.NodeType() == nodes.OutlineNodeType {
			g.linebuff.Writeln(g.colored(c_WHITE, "\n    Examples:"))
			g.linebuff.Flush()
			g.FormatTable(node.(nodes.OutlineNode).Examples().Table())
		}
	}
	g.linebuff.Flush()
}

func (g *gherkinPrettyPrinter) FormatStep(node nodes.StepNode) {
	g.formatStep(node)
	g.linebuff.Flush()
}

func (g *gherkinPrettyPrinter) coloredComment(node nodes.CommentNode) *styledString {
	if node != nil {
		str := node.Comment()
		if len(str) > 1 && str[0:1] != " " {
			str = " " + str
		}
		return g.colored(c_GRAY, "#%s", str)
	}
	return &styledString{}
}

func (g *gherkinPrettyPrinter) formatStep(node nodes.StepNode) {
	var stepTypeFmt string
	if g.gpf.CenterSteps {
		stepTypeFmt = "%9s"
	} else {
		stepTypeFmt = "    %s"
	}
	g.linebuff.Writeln(
		g.joinStyledStrings(
			g.colored(c_BOLD_GREEN, stepTypeFmt, node.StepType()),
			g.colored(c_GREEN, " %s", node.Text()),
		),
		g.coloredComment(node.Comment()),
	)

	if node.PyString() != nil {
		g.linebuff.Flush()
		g.FormatPyString(node.PyString())
	}

	if node.Table() != nil {
		g.linebuff.Flush()
		g.FormatTable(node.Table())
	}
}

func (g *gherkinPrettyPrinter) FormatTable(node nodes.TableNode) {
	rows := node.Rows()
	cellwidth := make(map[int]int, 100)
	for _, row := range rows {
		for c, str := range row {
			l := utf8.RuneCountInString(str)
			if l > cellwidth[c] {
				cellwidth[c] = l
			}
		}
	}
	wood := g.colored(c_BOLD_YELLOW, "|").String()
	for _, row := range rows {
		g.write("      ")
		for c, str := range row {
			_, err := strconv.ParseFloat(str, 64)
			var fmtStr string
			if err != nil {
				fmtStr = g.colored(c_YELLOW, fmt.Sprintf(" %%-%ds ", cellwidth[c])).String()
			} else {
				fmtStr = g.colored(c_YELLOW, fmt.Sprintf(" %%%ds ", cellwidth[c])).String()
			}
			g.write(wood + fmt.Sprintf(fmtStr, str))
		}
		g.write(wood + "\n")
	}
}

func (g *gherkinPrettyPrinter) FormatPyString(node nodes.PyStringNode) {
	prefix := "      "
	quotes := g.colored(c_BOLD, "\"\"\"").String()
	g.write(prefix + quotes + "\n")
	g.write(g.colored(c_YELLOW, prefixLines(prefix, node.String())).String())
	g.write(quotes + "\n")
}

func prefixLines(prefix, str string) string {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func fmtTags(tags []string) string {
	for i, tag := range tags {
		tags[i] = "@" + tag
	}
	return strings.Join(tags, " ")
}
