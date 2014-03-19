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
	c_BOLD_RED               = "31;1"
	c_BOLD_GREEN             = "32;1"
	c_BOLD_YELLOW            = "33;1"
	c_BOLD_BLUE              = "34;1"
	c_BOLD_MAGENTA           = "35;1"
	c_BOLD_CYAN              = "36;1"
)

func (g *gherkinPrettyPrinter) colored(color ansiStyle, str string) string {
	if g.gpf.AnsiColors {
		return fmt.Sprintf("\x1B[%sm%s\x1B[m", color, str)
	}
	return str
}

func (g *gherkinPrettyPrinter) FormatFeature(node FeatureNode) {
	tags := node.Tags()
	if len(tags) > 0 {
		g.write(g.colored(c_CYAN, fmtTags(tags)) + "\n")
	}
	g.write(fmt.Sprintf("%s: %s\n", g.colored(c_BOLD, node.NodeType().String()), node.Title()))
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
		g.write("  " + g.colored(c_CYAN, fmtTags(tags)) + "\n")
	}
	switch node.NodeType() {
	case BackgroundNodeType:
		g.write(fmt.Sprintf("  %s\n", g.colored(c_BOLD, "Background:")))
	case ScenarioNodeType:
		g.write(fmt.Sprintf("  %s %s\n", g.colored(c_BOLD, "Scenario:"), g.colored(c_WHITE, node.Title())))
	case OutlineNodeType:
		g.write(fmt.Sprintf("  %s %s\n", g.colored(c_BOLD, "Scenario Outline:"), g.colored(c_WHITE, node.Title())))
	}
	for _, step := range node.Steps() {
		g.FormatStep(step)
	}
	if node.NodeType() == OutlineNodeType {
		g.write(g.colored(c_WHITE, "\n    Examples:\n"))
		g.FormatTable(node.(OutlineNode).Examples().Table())
	}
}

func (g *gherkinPrettyPrinter) FormatStep(node StepNode) {
	if g.gpf.CenterSteps {
		g.write(g.colored(c_BOLD_GREEN, fmt.Sprintf("%9s", node.StepType())) + g.colored(c_GREEN, fmt.Sprintf(" %s\n", node.Text())))
	} else {
		g.write(fmt.Sprintf("    %s %s\n", g.colored(c_BOLD_GREEN, node.StepType()), g.colored(c_GREEN, node.Text())))
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
	wood := g.colored(c_BOLD_YELLOW, "|")
	for _, row := range rows {
		g.write("      ")
		for c, str := range row {
			_, err := strconv.ParseFloat(str, 64)
			var fmtStr string
			if err != nil {
				fmtStr = g.colored(c_YELLOW, fmt.Sprintf(" %%-%ds ", cellwidth[c]))
			} else {
				fmtStr = g.colored(c_YELLOW, fmt.Sprintf(" %%%ds ", cellwidth[c]))
			}
			g.write(wood + fmt.Sprintf(fmtStr, str))
		}
		g.write(wood + "\n")
	}
}

func (g *gherkinPrettyPrinter) FormatPyString(node PyStringNode) {
	prefix := "      "
	quotes := g.colored(c_BOLD, "\"\"\"")
	g.write(prefix + quotes + "\n")
	g.write(g.colored(c_YELLOW, prefixLines(prefix, node.String())))
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
