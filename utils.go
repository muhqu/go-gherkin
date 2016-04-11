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
	lines = omitSuccessiveEmptyLines(lines)
	return trimWS(strings.Join(lines, "\n"))
}

func omitSuccessiveEmptyLines(lines []string) []string {
	var out []string
	var b1, b2 bool
	for _, line := range lines {
		if line != "" {
			if b1 && !b2 {
				out = append(out, "")
			}
			out = append(out, line)
			b1 = true
			b2 = true
		} else {
			b2 = false
		}
	}
	return out
}

func trimNL(str string) string {
	return strings.Trim(str, "\r\n")
}
