package formater_test

import (
	"github.com/muhqu/go-gherkin"
	"github.com/muhqu/go-gherkin/formater"
	"io"
	"os"
)

var unformatedGherkin = `@dead @simple Feature: Dead Simple Calculator
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
Then the result should be 9
Scenario: Follow user actions
When I do the following user actions:
| action | key |
| key down | 2 |
| key up | 2 |
| key down | + |
| key up | + |
| key down | 4 |
| key up | 4 |
And I press the key "="
Then the result should be 6`

func ExampleGherkinPrettyFormater_1() {

	fmt := &formater.GherkinPrettyFormater{}

	// unformatedGherkin := `@dead @simple Feature: Dead Simple Calculator ...`
	gp := gherkin.NewGherkinDOMParser(unformatedGherkin)
	reader := fmt.Format(gp)
	io.Copy(os.Stdout, reader)

	// Output:
	// @dead @simple
	// Feature: Dead Simple Calculator
	//   Bla Bla
	//   Bla
	//
	//   Background:
	//     Given a Simple Calculator
	//
	//   @wip
	//   Scenario: Adding 2 numbers
	//     When I press the key "2"
	//     And I press the key "+"
	//     And I press the key "2"
	//     And I press the key "="
	//     Then the result should be 4
	//
	//   @wip @expensive
	//   Scenario Outline: Simple Math
	//     When I press the key "<left>"
	//     And I press the key "<operator>"
	//     And I press the key "<right>"
	//     And I press the key "="
	//     Then the result should be "<result>"
	//
	//     Examples:
	//       | left | operator | right | result |
	//       |    2 | +        |     2 |      4 |
	//       |    3 | +        |     4 |      7 |
	//
	//   Scenario: Adding 3 numbers
	//     When I press the following keys:
	//       """
	//         2
	//       + 2
	//       + 5
	//         =
	//       """
	//     Then the result should be 9
	//
	//   Scenario: Follow user actions
	//     When I do the following user actions:
	//       | action   | key |
	//       | key down |   2 |
	//       | key up   |   2 |
	//       | key down | +   |
	//       | key up   | +   |
	//       | key down |   4 |
	//       | key up   |   4 |
	//     And I press the key "="
	//     Then the result should be 6
	//
}

func ExampleGherkinPrettyFormater_2() {

	fmt := &formater.GherkinPrettyFormater{
		CenterSteps: true,
	}

	// unformatedGherkin := `@dead @simple Feature: Dead Simple Calculator ...`
	gp := gherkin.NewGherkinDOMParser(unformatedGherkin)

	reader := fmt.Format(gp)
	io.Copy(os.Stdout, reader)

	// Output:
	// @dead @simple
	// Feature: Dead Simple Calculator
	//   Bla Bla
	//   Bla
	//
	//   Background:
	//     Given a Simple Calculator
	//
	//   @wip
	//   Scenario: Adding 2 numbers
	//      When I press the key "2"
	//       And I press the key "+"
	//       And I press the key "2"
	//       And I press the key "="
	//      Then the result should be 4
	//
	//   @wip @expensive
	//   Scenario Outline: Simple Math
	//      When I press the key "<left>"
	//       And I press the key "<operator>"
	//       And I press the key "<right>"
	//       And I press the key "="
	//      Then the result should be "<result>"
	//
	//     Examples:
	//       | left | operator | right | result |
	//       |    2 | +        |     2 |      4 |
	//       |    3 | +        |     4 |      7 |
	//
	//   Scenario: Adding 3 numbers
	//      When I press the following keys:
	//       """
	//         2
	//       + 2
	//       + 5
	//         =
	//       """
	//      Then the result should be 9
	//
	//   Scenario: Follow user actions
	//      When I do the following user actions:
	//       | action   | key |
	//       | key down |   2 |
	//       | key up   |   2 |
	//       | key down | +   |
	//       | key up   | +   |
	//       | key down |   4 |
	//       | key up   |   4 |
	//       And I press the key "="
	//      Then the result should be 6
	//
}
