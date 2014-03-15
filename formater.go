package gherkin

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

type GherkinFormater interface {
	Format(GherkinDOM) io.Reader
}

func (g *gherkinDOMParser) Format(f GherkinFormater) io.Reader {
	return f.Format(g)
}

// -------------------

type GherkinPrettyFormater struct {
	AnsiColors  bool
	CenterSteps bool
}
type gherkinPrettyPrinter struct {
	gpf *GherkinPrettyFormater
	io.Writer
}

func (gpf *GherkinPrettyFormater) Format(gd GherkinDOM) io.Reader {
	sw := new(bytes.Buffer)
	g := &gherkinPrettyPrinter{
		gpf:    gpf,
		Writer: sw,
	}
	g.FormatFeature(gd.Feature())
	return bytes.NewReader(sw.Bytes())
}

func (g *gherkinPrettyPrinter) write(s string) {
	io.WriteString(g.Writer, s)
}

type AnsiStyle string

const (
	BOLD         AnsiStyle = "1"
	WHITE                  = "29"
	BLACK                  = "30"
	RED                    = "31"
	GREEN                  = "32"
	YELLOW                 = "33"
	BLUE                   = "34"
	MAGENTA                = "35"
	CYAN                   = "36"
	BOLD_RED               = "31;1"
	BOLD_GREEN             = "32;1"
	BOLD_YELLOW            = "33;1"
	BOLD_BLUE              = "34;1"
	BOLD_MAGENTA           = "35;1"
	BOLD_CYAN              = "36;1"
)

func (g *gherkinPrettyPrinter) colored(color AnsiStyle, str string) string {
	if g.gpf.AnsiColors {
		return fmt.Sprintf("\x1B[%sm%s\x1B[m", color, str)
	}
	return str
}

func (g *gherkinPrettyPrinter) FormatFeature(node FeatureNode) {
	tags := node.Tags()
	if len(tags) > 0 {
		g.write(g.colored(CYAN, fmtTags(tags)) + "\n")
	}
	g.write(fmt.Sprintf("%s: %s\n", g.colored(BOLD, node.NodeType().String()), node.Title()))
	if node.Description() != "" {
		g.write(prefixLines("  ", node.Description()) + "\n")
	}
	g.write("\n")

	if node.Background() != nil {
		g.FormatScenario(node.Background())
		g.write("\n")
	}

	for _, scenario := range node.Scenarios() {
		g.FormatScenario(scenario)
		g.write("\n")
	}
}

func (g *gherkinPrettyPrinter) FormatScenario(node ScenarioNode) {
	tags := node.Tags()
	if len(tags) > 0 {
		g.write("  " + g.colored(CYAN, fmtTags(tags)) + "\n")
	}
	switch node.NodeType() {
	case BackgroundNodeType:
		g.write(fmt.Sprintf("  %s\n", g.colored(BOLD, "Background:")))
	case ScenarioNodeType:
		g.write(fmt.Sprintf("  %s %s\n", g.colored(BOLD, "Scenario:"), g.colored(WHITE, node.Title())))
	case OutlineNodeType:
		g.write(fmt.Sprintf("  %s %s\n", g.colored(BOLD, "Scenario Outline:"), g.colored(WHITE, node.Title())))
	}
	for _, step := range node.Steps() {
		g.FormatStep(step)
	}
	if node.NodeType() == OutlineNodeType {
		g.write(g.colored(WHITE, "\n    Examples:\n"))
		g.FormatTable(node.(OutlineNode).Examples().Table())
	}
}

func (g *gherkinPrettyPrinter) FormatStep(node StepNode) {
	if g.gpf.CenterSteps {
		g.write(g.colored(BOLD_GREEN, fmt.Sprintf("%9s", node.StepType())) + g.colored(GREEN, fmt.Sprintf(" %s\n", node.Text())))
	} else {
		g.write(fmt.Sprintf("    %s %s\n", g.colored(BOLD_GREEN, node.StepType()), g.colored(GREEN, node.Text())))
	}

	if node.PyString() != nil {
		g.FormatPyString(node.PyString())
	}

	if node.Table() != nil {
		g.FormatTable(node.Table())
	}
}

func (g *gherkinPrettyPrinter) FormatTable(node TableNode) {
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
	wood := g.colored(BOLD_YELLOW, "|")
	for _, row := range rows {
		g.write("      ")
		for c, str := range row {
			_, err := strconv.ParseFloat(str, 64)
			var fmtStr string
			if err != nil {
				fmtStr = g.colored(YELLOW, fmt.Sprintf(" %%-%ds ", cellwidth[c]))
			} else {
				fmtStr = g.colored(YELLOW, fmt.Sprintf(" %%%ds ", cellwidth[c]))
			}
			g.write(wood + fmt.Sprintf(fmtStr, str))
		}
		g.write(wood + "\n")
	}
}

func (g *gherkinPrettyPrinter) FormatPyString(node PyStringNode) {
	prefix := "      "
	quotes := g.colored(BOLD, "\"\"\"")
	g.write(prefix + quotes + "\n")
	g.write(g.colored(YELLOW, prefixLines(prefix, node.String())))
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
