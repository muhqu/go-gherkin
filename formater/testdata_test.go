package formater_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/muhqu/go-gherkin"
	"github.com/muhqu/go-gherkin/formater"
	"github.com/muhqu/gomega-additions/unifieddiff"
	"github.com/onsi/gomega"
)

func TestData(t *testing.T) {
	gomega.RegisterTestingT(t)

	re := regexp.MustCompile(`(.*)\.input\..*\.feature`)

	matches, err := filepath.Glob("testdata/*.input.*.feature")
	if err != nil {
		t.Fatal(err)
	}

	gfmt := &formater.GherkinPrettyFormater{}

	for _, inputFilename := range matches {
		t.Logf("inputFilename: %v", inputFilename)
		inputBytes, err := ioutil.ReadFile(inputFilename)
		if err != nil {
			t.Errorf("error reading file %q: %s", inputFilename, err)
			continue
		}

		formatedBytes, err := FormatBytes(gfmt, inputBytes)
		if err != nil {
			t.Errorf("error formating %q: %s", inputFilename, err)
			continue
		}

		expectedFilename := re.ReplaceAllString(inputFilename, `${1}.expected.feature`)
		expectedBytes, err := ioutil.ReadFile(expectedFilename)
		if err == nil {
			t.Logf("expected file: %q", expectedFilename)
			if !CheckEqual(t, fmt.Sprintf("invalid formating: %q", inputFilename), formatedBytes, expectedBytes) {
				continue
			}

			formated2ndBytes, err := FormatBytes(gfmt, formatedBytes)
			if err != nil {
				t.Errorf("error formating (2nd run) %q: %s", inputFilename, err)
				continue
			}

			if !CheckEqual(t, fmt.Sprintf("invalid formating (2nd run) : %q", inputFilename), formated2ndBytes, expectedBytes) {
				continue
			}

		} else if os.IsNotExist(err) {
			t.Logf("expected file does not exist: %q", expectedFilename)
			err = ioutil.WriteFile(expectedFilename, formatedBytes, os.FileMode(0644))
			if err != nil {
				t.Errorf("error writing file %q: %s", expectedFilename, err)
				continue
			}
		}
	}
}

func CheckEqual(t *testing.T, msg string, expected, actual []byte) bool {
	if bytes.Equal(expected, actual) {
		return true
	}
	t.Error(msg)
	m := unifieddiff.MatchString(string(expected))
	m.Match(string(actual))
	t.Error(m.FailureMessage(""))
	return false
}

func FormatBytes(gfmt *formater.GherkinPrettyFormater, in []byte) (out []byte, err error) {
	domparser := gherkin.NewGherkinDOMParser(string(in))
	dom, err := domparser.ParseDOM()
	if err == nil {
		buf := new(bytes.Buffer)
		gfmt.Format(dom, buf)
		out = buf.Bytes()
	}
	return
}
