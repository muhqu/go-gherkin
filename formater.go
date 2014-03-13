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
	FormatGherkin(GherkinDOM) io.Reader
}

func (g *gherkinDOMParser) Format(f GherkinFormater) io.Reader {
	return f.FormatGherkin(g)
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

func (gpf *GherkinPrettyFormater) FormatGherkin(gd GherkinDOM) io.Reader {
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

func (g *gherkinPrettyPrinter) FormatFeature(node FeatureNode) {
	tags := node.Tags()
	if len(tags) > 0 {
		g.write(fmtTags(tags) + "\n")
	}
	g.write(fmt.Sprintf("%s: %s\n", node.NodeType().String(), node.Title()))
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
		g.write("  " + fmtTags(tags) + "\n")
	}
	switch node.NodeType() {
	case BackgroundNodeType:
		g.write(fmt.Sprintf("  %s:\n", "Background"))
	case ScenarioNodeType:
		g.write(fmt.Sprintf("  %s: %s\n", "Scenario", node.Title()))
	case OutlineNodeType:
		g.write(fmt.Sprintf("  %s: %s\n", "Scenario Outline", node.Title()))
	}
	for _, step := range node.Steps() {
		g.FormatStep(step)
	}
	if node.NodeType() == OutlineNodeType {
		g.write("\n    Examples:\n")
		g.FormatTable(node.(OutlineNode).Examples().Table())
	}
}

func (g *gherkinPrettyPrinter) FormatStep(node StepNode) {
	if g.gpf.CenterSteps {
		g.write(fmt.Sprintf("%9s %s\n", node.StepType(), node.Text()))
	} else {
		g.write(fmt.Sprintf("    %s %s\n", node.StepType(), node.Text()))
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
	for _, row := range rows {
		g.write("      ")
		for c, str := range row {
			_, err := strconv.ParseFloat(str, 64)
			var fmtStr string
			if err != nil {
				fmtStr = fmt.Sprintf("| %%-%ds ", cellwidth[c])
			} else {
				fmtStr = fmt.Sprintf("| %%%ds ", cellwidth[c])
			}
			g.write(fmt.Sprintf(fmtStr, str))
		}
		g.write("|\n")
	}
}

func (g *gherkinPrettyPrinter) FormatPyString(node PyStringNode) {
	prefix := "      "
	g.write(prefix + "\"\"\"\n")
	g.write(prefixLines(prefix, node.String()))
	g.write("\"\"\"\n")
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
