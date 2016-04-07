package gherkin_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/muhqu/go-gherkin"
	"github.com/muhqu/go-gherkin/nodes"
	"github.com/stretchr/testify/assert"
)

func prettyPrintingLogFn(logPrefix string) func(msg string, args ...interface{}) {
	depth := 0
	return func(msg string, args ...interface{}) {
		isBegin := msg[0:5] == "Begin"
		isEnd := msg[0:3] == "End"
		if isEnd {
			depth = depth - 1
			if depth < 0 {
				depth = 0
			}
		}
		line := fmt.Sprintf(msg, args...)
		depthPrefix := strings.Repeat("  ", depth)
		fmt.Printf("%s%s%s\n", logPrefix, depthPrefix, line)
		if isBegin {
			depth = depth + 1
		}
	}
}

func parse(t *testing.T, logPrefix, text string) (gherkin.GherkinDOMParser, error) {
	gp := gherkin.NewGherkinDOMParser(text)
	if logPrefix != "" {
		gp.WithLogFn(prettyPrintingLogFn(logPrefix))
	}
	gp.Init()
	err := gp.Parse()
	if err == nil {
		gp.Execute()
	}
	return gp, err
}

func mustDomParse(t *testing.T, logPrefix, text string) gherkin.GherkinDOMParser {
	gp, err := parse(t, logPrefix, text)
	assert.NoError(t, err)
	if err != nil {
		t.FailNow()
	}
	return gp
}

func verifyDeadSimpleCalculator(t *testing.T, logPrefix, text string) {
	gp := mustDomParse(t, logPrefix, text)

	feature := gp.Feature()
	assert.NotNil(t, feature)
	assert.Equal(t, "Dead Simple Calculator", feature.Title())
	assert.Equal(t, "Bla Bla\nBla", feature.Description())
	assert.Equal(t, 3, len(feature.Scenarios()), "Number of scenarios")
	assert.Equal(t, 2, len(feature.Tags()), "Number of tags")
	assert.Equal(t, []string{"dead", "simple"}, feature.Tags(), "Feature Tags")

	assert.NotNil(t, feature.Background())
	assert.Equal(t, 1, len(feature.Background().Steps()), "Number of background steps")
	assert.Equal(t, "Given", feature.Background().Steps()[0].StepType())
	assert.Equal(t, "a Simple Calculator", feature.Background().Steps()[0].Text())

	scenario1 := feature.Scenarios()[0]
	assert.NotNil(t, scenario1)
	assert.Equal(t, nodes.ScenarioNodeType, scenario1.NodeType())
	assert.Equal(t, 1, len(scenario1.Tags()), "Number of tags on Scenario 1")
	assert.Equal(t, []string{"wip"}, scenario1.Tags(), "Tags on Senario 1")
	assert.Equal(t, 5, len(scenario1.Steps()), "Number of steps in Scenario 1")
	assert.Equal(t, "When", scenario1.Steps()[0].StepType())
	assert.Equal(t, "I press the key \"2\"", scenario1.Steps()[0].Text())

	scenario2 := feature.Scenarios()[1]
	assert.NotNil(t, scenario2)
	assert.Equal(t, nodes.OutlineNodeType, scenario2.NodeType())
	scenario2o, ok := scenario2.(nodes.OutlineNode)
	assert.True(t, ok)
	assert.Equal(t, 2, len(scenario2.Tags()), "Number of tags on Scenario 2")
	assert.Equal(t, []string{"wip", "expensive"}, scenario2.Tags(), "Tags on Senario 2")
	assert.Equal(t, 5, len(scenario2.Steps()), "Number of steps in Scenario 2")
	assert.Equal(t, "When", scenario2.Steps()[0].StepType())
	assert.Equal(t, "I press the key \"<left>\"", scenario2.Steps()[0].Text())
	assert.Equal(t, [][]string{
		{"left", "operator", "right", "result"},
		{"2", "+", "2", "4"},
		{"3", "+", "4", "7"},
	}, scenario2o.Examples().Table().Rows())

	scenario3 := feature.Scenarios()[2]
	assert.NotNil(t, scenario3)
	assert.Equal(t, nodes.ScenarioNodeType, scenario3.NodeType())
	assert.Equal(t, 0, len(scenario3.Tags()), "Number of tags on Scenario 3")
	assert.Equal(t, 2, len(scenario3.Steps()), "Number of steps in Scenario 3")
	assert.Equal(t, "When", scenario3.Steps()[0].StepType())
	assert.Equal(t, "I press the following keys:", scenario3.Steps()[0].Text())
	assert.NotNil(t, scenario3.Steps()[0].PyString())
	assert.Equal(t, "  2\n+ 2\n+ 5\n  =\n", scenario3.Steps()[0].PyString().String())
}

const benchmarkGherkinText = `
@dead @simple
Feature: Dead Simple Calculator
  Bla Bla
  Bla

  Background:
    Given a Simple Calculator

  @wip
  Scenario: Adding 2 numbers
     When I press the key "2"
      And I press the key "+"
	  And I press the key "2"
      And I press the key "="
     Then the result should be 4

  @wip @expensive
  Scenario Outline: Simple Math
     When I press the key "<left>"
      And I press the key "<operator>"
	  And I press the key "<right>"
      And I press the key "="
     Then the result should be "<result>"

    Examples:
     | left   | operator | right   | result |
     | 2      | +        | 2       | 4      |
     | 3      | +        | 4       | 7      |

  Scenario: Adding 3 numbers
     When I press the following keys:
     """
       2
     + 2
     + 5
       =
     """
     Then the result should be 9

`

func Benchmark_NewGherkinDOMParser(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	for i := 0; i < b.N; i++ { // use b.N for looping
		gherkin.NewGherkinDOMParser(benchmarkGherkinText)
	}
}
func Benchmark_NewGherkinDOMParserAndParse(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	for i := 0; i < b.N; i++ { // use b.N for looping
		gp := gherkin.NewGherkinDOMParser(benchmarkGherkinText)
		gp.Init()
		gp.Parse()
	}
}
func Benchmark_JustReParseWithCachedGherkinDOMParser(b *testing.B) { //benchmark function starts with "Benchmark" and takes a pointer to type testing.B
	b.StopTimer()
	gp := gherkin.NewGherkinDOMParser(benchmarkGherkinText)
	b.StartTimer()
	for i := 0; i < b.N; i++ { // use b.N for looping
		gp.Init()
		gp.Parse()
	}
}

func TestParsingRegular(t *testing.T) {
	verifyDeadSimpleCalculator(t, "", `
@dead @simple
Feature: Dead Simple Calculator
  Bla Bla
  Bla

  Background:
    Given a Simple Calculator

  @wip
  Scenario: Adding 2 numbers
     When I press the key "2"
      And I press the key "+"
	  And I press the key "2"
      And I press the key "="
     Then the result should be 4

  @wip @expensive
  Scenario Outline: Simple Math
     When I press the key "<left>"
      And I press the key "<operator>"
	  And I press the key "<right>"
      And I press the key "="
     Then the result should be "<result>"

    Examples:
     | left   | operator | right   | result |
     | 2      | +        | 2       | 4      |
     | 3      | +        | 4       | 7      |

  Scenario: Adding 3 numbers
     When I press the following keys:
     """
       2
     + 2
     + 5
       =
     """
     Then the result should be 9

`)
}

func TestParsingTabAligned(t *testing.T) {
	verifyDeadSimpleCalculator(t, "", `
@dead @simple
Feature: Dead Simple Calculator
	Bla Bla
	Bla

Background:
	Given a Simple Calculator

@wip
Scenario: Adding 2 numbers
	When I press the key "2"
	And I press the key "+"
	And I press the key "2"
	And I press the key "="
	Then the result should be 4

@wip @expensive
Scenario Outline: Simple Math
	When I press the key "<left>"
	And I press the key "<operator>"
	And I press the key "<right>"
	And I press the key "="
	Then the result should be "<result>"

Examples:
	| left   | operator | right   | result |
	| 2      | +        | 2       | 4      |
	| 3      | +        | 4       | 7      |

Scenario: Adding 3 numbers
	When I press the following keys:
	"""
	  2
	+ 2
	+ 5
	  =
	"""
	Then the result should be 9
`)
}

func TestParsingCondensedAndTrailingWhitespace(t *testing.T) {
	verifyDeadSimpleCalculator(t, "", `@dead @simple Feature: Dead Simple Calculator         
Bla Bla                                 
Bla                                     
Background:                             
Given a Simple Calculator               
@wip Scenario: Adding 2 numbers              
When I press the key "2"                
And I press the key "+"                 
And I press the key "2"                 
And I press the key "="                 
Then the result should be 4             
@wip @expensive Scenario Outline: Simple Math           
When I press the key "<left>"           
And I press the key "<operator>"        
And I press the key "<right>"           
And I press the key "="                 
Then the result should be "<result>"    
Examples:                               
| left | operator | right | result |    
| 2 | + | 2 | 4 |                       
| 3 | + | 4 | 7 |                       
Scenario: Adding 3 numbers              
When I press the following keys:        
"""
  2
+ 2
+ 5
  =
"""
Then the result should be 9`)
}

func TestParsingMinimalNoScenarios(t *testing.T) {
	gp := mustDomParse(t, "", `Feature: Hello World`)
	feature := gp.Feature()
	assert.NotNil(t, feature)
	assert.Equal(t, "Hello World", feature.Title())
}

func TestParsingMinimalNoSteps(t *testing.T) {
	gp := mustDomParse(t, "", `Feature: Hello World
Scenario: Nice people`)
	feature := gp.Feature()
	assert.NotNil(t, feature)
	assert.Equal(t, "Hello World", feature.Title())

	assert.Equal(t, 1, len(feature.Scenarios()), "Number of Scenarios")

	scenario1 := feature.Scenarios()[0]
	assert.NotNil(t, scenario1)
	assert.Equal(t, nodes.ScenarioNodeType, scenario1.NodeType())
	assert.Equal(t, 0, len(scenario1.Steps()), "Number of steps in Scenario 1")
}

func TestParsingMinimalWithSteps(t *testing.T) {
	gp := mustDomParse(t, "", `
Feature: Hello World

  Scenario: Nice people
    Given a nice person called "Bob"
      And a nice person called "Lisa"
     When "Bob" says to "Lisa": "Hello!"
     Then "Lisa" should reply to "Bob": "Hello!"

`)
	feature := gp.Feature()
	assert.NotNil(t, feature)
	assert.Equal(t, "Hello World", feature.Title())

	assert.Equal(t, 1, len(feature.Scenarios()), "Number of Scenarios")

	scenario1 := feature.Scenarios()[0]
	assert.NotNil(t, scenario1)
	assert.Equal(t, nodes.ScenarioNodeType, scenario1.NodeType())
	assert.Equal(t, 4, len(scenario1.Steps()), "Number of steps in Scenario 1")
	i := 0
	assert.Equal(t, "Given", scenario1.Steps()[i].StepType())
	assert.Equal(t, `a nice person called "Bob"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "And", scenario1.Steps()[i].StepType())
	assert.Equal(t, `a nice person called "Lisa"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "When", scenario1.Steps()[i].StepType())
	assert.Equal(t, `"Bob" says to "Lisa": "Hello!"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "Then", scenario1.Steps()[i].StepType())
	assert.Equal(t, `"Lisa" should reply to "Bob": "Hello!"`, scenario1.Steps()[i].Text())
}

func TestParsingFailure(t *testing.T) {
	_, err := parse(t, "", `
Feature: Dead Simple Calculator
  Scenario:
    Hurtz
`)
	assert.Error(t, err)
}

func TestParsingSimpleComments(t *testing.T) {
	gp := mustDomParse(t, "", `
Feature: Hello World                             # feature comment

  Scenario: Nice people                          # scenario 1 comment
    Given a nice person called "Bob"             # step 1.1 comment
      And a nice person called "Lisa"            # step 1.2 comment
     When "Bob" says to "Lisa": "Hello!"         # step 1.3 comment
     Then "Lisa" should reply to "Bob": "Hello!" # step 1.4 comment`)
	feature := gp.Feature()
	if ok := assert.NotNil(t, feature); !ok {
		return
	}
	assert.Equal(t, "Hello World", feature.Title())

	if ok := assert.Equal(t, 1, len(feature.Scenarios()), "Number of Scenarios"); !ok {
		return
	}

	scenario1 := feature.Scenarios()[0]
	if ok := assert.NotNil(t, scenario1); !ok {
		return
	}
	assert.Equal(t, nodes.ScenarioNodeType, scenario1.NodeType())
	assert.Equal(t, 4, len(scenario1.Steps()), "Number of steps in Scenario 1")
	i := 0
	assert.Equal(t, "Given", scenario1.Steps()[i].StepType())
	assert.Equal(t, `a nice person called "Bob"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "And", scenario1.Steps()[i].StepType())
	assert.Equal(t, `a nice person called "Lisa"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "When", scenario1.Steps()[i].StepType())
	assert.Equal(t, `"Bob" says to "Lisa": "Hello!"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "Then", scenario1.Steps()[i].StepType())
	assert.Equal(t, `"Lisa" should reply to "Bob": "Hello!"`, scenario1.Steps()[i].Text())
}

func TestParsingDifficultComments(t *testing.T) {
	gp := mustDomParse(t, "", `
@awesome @dude                                           # feature tag comment
Feature: Hello "#World"                                  # feature comment
  Bla bla                                                # feature description comment
  Bla                                                    # feature description comment
                                                         # blank line comment
  @wip @wop                                              # scenario tag comment
  Scenario: Nice people                                  # scenario 1 comment
    Given a nice person called "#Bob"                    # step 1.1 comment
      And a nice person called "Lisa#\"Bang"             # step 1.2 comment
     When "#Bob" says to "Lisa#\"Bang": "Hello!"         # step 1.3 comment
     Then "Lisa#\"Bang" should reply to "#Bob": "Hello!" # step 1.4 comment
                                                         # blank line comment
`)
	feature := gp.Feature()
	if ok := assert.NotNil(t, feature); !ok {
		return
	}
	assert.Equal(t, `Hello "#World"`, feature.Title())
	assert.Equal(t, "Bla bla\nBla", feature.Description())
	assert.Equal(t, []string{"awesome", "dude"}, feature.Tags())

	if ok := assert.Equal(t, 1, len(feature.Scenarios()), "Number of Scenarios"); !ok {
		return
	}

	scenario1 := feature.Scenarios()[0]
	if ok := assert.NotNil(t, scenario1); !ok {
		return
	}
	assert.Equal(t, nodes.ScenarioNodeType, scenario1.NodeType())
	assert.Equal(t, []string{"wip", "wop"}, scenario1.Tags())
	assert.Equal(t, 4, len(scenario1.Steps()), "Number of steps in Scenario 1")
	i := 0
	assert.Equal(t, "Given", scenario1.Steps()[i].StepType())
	assert.Equal(t, `a nice person called "#Bob"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "And", scenario1.Steps()[i].StepType())
	assert.Equal(t, `a nice person called "Lisa#\"Bang"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "When", scenario1.Steps()[i].StepType())
	assert.Equal(t, `"#Bob" says to "Lisa#\"Bang": "Hello!"`, scenario1.Steps()[i].Text())
	i += 1
	assert.Equal(t, "Then", scenario1.Steps()[i].StepType())
	assert.Equal(t, `"Lisa#\"Bang" should reply to "#Bob": "Hello!"`, scenario1.Steps()[i].Text())
}

func TestParsingBlankQuotedStrings(t *testing.T) {
	gp := mustDomParse(t, "", `
Feature: Empty "" Quotes

  Scenario: Allow ""
    When "" is present
`)

	feature := gp.Feature()
	if ok := assert.NotNil(t, feature); !ok {
		return
	}

	assert.Equal(t, `Empty "" Quotes`, feature.Title())

	if ok := assert.Equal(t, 1, len(feature.Scenarios()), "Number of Scenarios"); !ok {
		return
	}

	scenario1 := feature.Scenarios()[0]
	if ok := assert.NotNil(t, scenario1); !ok {
		return
	}

	assert.Equal(t, `Allow ""`, scenario1.Title())

	assert.Equal(t, 1, len(scenario1.Steps()), "Number of steps in Scenario 1")
	i := 0
	assert.Equal(t, "When", scenario1.Steps()[i].StepType())
	assert.Equal(t, `"" is present`, scenario1.Steps()[i].Text())
}

func TestParsingDescriptionWithBlankLines(t *testing.T) {
	gp := mustDomParse(t, "", `
Feature: Description With Blank Lines

  Whitespace should be allowed.

  (Multiple times)
`)

	feature := gp.Feature()
	if ok := assert.NotNil(t, feature); !ok {
		return
	}

	assert.Equal(t, "Whitespace should be allowed.\n\n(Multiple times)", feature.Description())
}

func TestParsingMultipleExamples(t *testing.T) {
	gp := mustDomParse(t, "", `
Feature: Account withdrawal

  Scenario Outline: Withdraw fixed amount
    Given I have <Balance> in my account
    When I choose to withdraw the fixed amount of <Withdrawal>
    Then I should <Outcome>
    And the balance of my account should be <Remaining>

    Examples:
      | Balance | Withdrawal | Outcome | Remaining |
      | $500 | $50 | receive $50 cash | $450 |
      | $500 | $100 | receive $100 cash | $400 |

    Examples:
      | Balance | Withdrawal | Outcome | Remaining |
      | $100 | $200 | see an error message | $100 |
      | $0 | $50 | see an error message | $0 |
`)

	feature := gp.Feature()
	if ok := assert.NotNil(t, feature); !ok {
		return
	}

	assert.Equal(t, "Account withdrawal", feature.Title())
}
