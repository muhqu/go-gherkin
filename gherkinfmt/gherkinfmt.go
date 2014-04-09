/*

Command gherkinfmt is a command-line gherkin formater and pretty printer.

Usage

	gherkinfmt OPTIONS
	  -centersteps[=false]   formating option, to control step alignment
	  -[no]color             explicitly enable/disable colors
	  -in PATH               path to input file, defaults to stdin
	  -out PATH              path to output file, defaults to stdout

Examples

	$ gherkinfmt -in path/to/some.feature

	$ cat path/to/some.feature | gherkinfmt -centersteps

*/
package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/muhqu/go-gherkin"
	"github.com/muhqu/go-gherkin/formater"
	"github.com/pebbe/util"
)

var err error
var colors bool
var colorsYes bool
var colorsNo bool
var centerSteps bool
var skipSteps bool
var verbose bool
var inputPath string
var inputReader io.Reader
var outputPath string
var outputWriter io.Writer

func initFlags() {
	flag.Usage = func() {
		self := path.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, `Usage: %[1]s OPTIONS

  -centersteps[=false]   formating option, to control step alignment
  -nosteps[=false]       omit steps, just print scenario headlines
  
  -[no]color             explicitly enable/disable colors
  -in PATH               path to input file, defaults to stdin
  -out PATH              path to output file, defaults to stdout
  -v                     more verbose error messages

Examples:
  
  $ %[1]s -in path/to/some.feature

  $ cat path/to/some.feature | %[1]s -centersteps

`, self)
	}
	flag.BoolVar(&colorsYes, "color", false, "explicitly enable colors")
	flag.BoolVar(&colorsNo, "nocolor", false, "explicitly disable colors")
	flag.BoolVar(&centerSteps, "centersteps", false, "formating option, to control step alignment")
	flag.BoolVar(&skipSteps, "nosteps", false, "omit steps, just print scenario headlines")
	flag.BoolVar(&verbose, "v", false, "more verbose error messages")
	flag.StringVar(&inputPath, "in", "", "path to input file")
	flag.StringVar(&outputPath, "out", "", "path to output file")
	flag.Parse()
}

func usageErr(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n       Use -h for help.\n", err.Error())
}

func usageErrWithVerboseHint(err error) {
	if verbose {
		usageErr(err)
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n       Use -h for help or use -v to increase verbosity\n", err.Error())
	}
}

func main() {
	initFlags()

	if inputPath != "" {
		inputReader, err = os.Open(inputPath)
		if err != nil {
			usageErr(err)
			return
		}
	} else {
		st, err := os.Stdin.Stat()
		if err != nil {
			usageErr(err)
			return
		}
		if st.Size() > 0 {
			inputReader = os.Stdin
		}
	}
	if inputReader == nil {
		usageErr(fmt.Errorf("Missing input (stdin OR -in flag)"))
		return
	}

	if outputPath != "" {
		outputWriter, err = os.Create(outputPath)
		if err != nil {
			usageErr(err)
			return
		}
	} else {
		outputWriter = os.Stdout
	}

	if colorsYes {
		colors = true
	} else if colorsNo {
		colors = false
	} else if outputWriter == os.Stdout {
		colors = util.IsTerminal(os.Stdout)
	}

	fmtr := &formater.GherkinPrettyFormater{
		AnsiColors:  colors,
		CenterSteps: centerSteps,
		SkipSteps:   skipSteps,
	}

	bytes, _ := ioutil.ReadAll(inputReader)
	content := string(bytes)
	gp := gherkin.NewGherkinDOMParser(content)
	gp.Init()
	err = gp.Parse()
	if err != nil {
		usageErrWithVerboseHint(fmt.Errorf("Parsing failed. invalid gherkin"))
		if verbose {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	fmtr.Format(gp, outputWriter)
}
