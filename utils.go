package gherkin

import (
	"strings"
)

func trimWS(str string) string {
	return strings.Trim(str, " \t\r\n")
}

func trimLeadingWS(str string) string {
	return strings.TrimLeft(str, " \t\r\n")
}

func trimWSML(str string) string {
	lines := strings.Split(str, "\n")
	for i, line := range lines {
		lines[i] = trimWS(line)
	}
	return trimWS(strings.Join(lines, "\n"))
}

func trimNL(str string) string {
	return strings.Trim(str, "\r\n")
}
