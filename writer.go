package gherkin

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type GherkinDOMWriter interface {
	Write(w io.Writer)
}

func (g *gherkinDOMParser) Write(w io.Writer) {
	g.Feature()
	domWrite(w, g.feature)
}

func domWrite(w io.Writer, n NodeInterface) {
	switch node := n.(type) {

	default:
		io.WriteString(w, fmt.Sprintf("unexpexted %T\n", node))

	case *featureNode:
		tags := node.Tags()
		if len(tags) > 0 {
			io.WriteString(w, fmtTags(tags)+"\n")
		}
		io.WriteString(w, fmt.Sprintf("Feature: %s\n", node.Title()))
		if node.Description() != "" {
			io.WriteString(w, prefixLines("  ", node.Description())+"\n")
		}
		io.WriteString(w, "\n")
		if node.background != nil {
			domWrite(w, node.background)
			io.WriteString(w, "\n")
		}
		for _, scenario := range node.scenarios {
			domWrite(w, scenario)
			io.WriteString(w, "\n")
		}

	case *outlineNode:
		tags := node.Tags()
		if len(tags) > 0 {
			io.WriteString(w, "  "+fmtTags(tags)+"\n")
		}
		io.WriteString(w, fmt.Sprintf("  Scenario Outline: %s\n", node.Title()))
		for _, step := range node.steps {
			domWrite(w, step)
		}
		io.WriteString(w, "    Examples:\n")
		domWrite(w, node.examples.table)

	case *backgroundNode:
		tags := node.Tags()
		if len(tags) > 0 {
			io.WriteString(w, "  "+fmtTags(tags)+"\n")
		}
		io.WriteString(w, "  Background:\n")
		for _, step := range node.steps {
			domWrite(w, step)
		}

	case *scenarioNode:
		tags := node.Tags()
		if len(tags) > 0 {
			io.WriteString(w, "  "+fmtTags(tags)+"\n")
		}
		io.WriteString(w, fmt.Sprintf("  Scenario: %s\n", node.Title()))
		for _, step := range node.steps {
			domWrite(w, step)
		}

	case *stepNode:
		io.WriteString(w, fmt.Sprintf("%9s %s\n", node.StepType(), node.Text()))
		if node.pyString != nil {
			domWrite(w, node.pyString)
		}
		if node.table != nil {
			domWrite(w, node.table)
		}

	case *pyStringNode:
		prefix := "      "
		io.WriteString(w, prefix+"\"\"\"\n")
		io.WriteString(w, prefixLines(prefix, node.String()))
		io.WriteString(w, "\"\"\"\n")

	case *tableNode:
		rows := node.Rows()
		cellwidth := make(map[int]int, 100)
		for _, row := range rows {
			for c, str := range row {
				l := len(str)
				if l > cellwidth[c] {
					cellwidth[c] = l
				}
			}
		}
		for _, row := range rows {
			io.WriteString(w, "      ")
			for c, str := range row {
				_, err := strconv.ParseFloat(str, 64)
				var fmtStr string
				if err != nil {
					fmtStr = fmt.Sprintf("| %%-%ds ", cellwidth[c])
				} else {
					fmtStr = fmt.Sprintf("| %%%ds ", cellwidth[c])
				}
				io.WriteString(w, fmt.Sprintf(fmtStr, str))
			}
			io.WriteString(w, "|\n")
		}
	}
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
