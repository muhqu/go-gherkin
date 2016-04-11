package formater_test

import (
	"os"

	"github.com/muhqu/go-gherkin"
	"github.com/muhqu/go-gherkin/formater"
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
	fmt.Format(gp, os.Stdout)

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

	fmt.Format(gp, os.Stdout)

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

func ExampleGherkinPrettyFormater_3() {

	fmt := &formater.GherkinPrettyFormater{
		SkipSteps: true,
	}

	// unformatedGherkin := `@dead @simple Feature: Dead Simple Calculator ...`
	gp := gherkin.NewGherkinDOMParser(unformatedGherkin)

	fmt.Format(gp, os.Stdout)

	// Output:
	// @dead @simple
	// Feature: Dead Simple Calculator
	//   Bla Bla
	//   Bla
	//
	//   @wip
	//   Scenario: Adding 2 numbers
	//
	//   @wip @expensive
	//   Scenario Outline: Simple Math
	//
	//   Scenario: Adding 3 numbers
	//
	//   Scenario: Follow user actions
	//
}

const unformatedGherkinWithMultipleExamples = `
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
`

func ExampleGherkinPrettyFormater_4() {

	fmt := &formater.GherkinPrettyFormater{}

	// unformatedGherkinWithMultipleExamples := `Feature: Account withdrawal ...`
	gp := gherkin.NewGherkinDOMParser(unformatedGherkinWithMultipleExamples)

	fmt.Format(gp, os.Stdout)

	// Output:
	//
	// Feature: Account withdrawal
	//
	//   Scenario Outline: Withdraw fixed amount
	//     Given I have <Balance> in my account
	//     When I choose to withdraw the fixed amount of <Withdrawal>
	//     Then I should <Outcome>
	//     And the balance of my account should be <Remaining>
	//
	//     Examples:
	//       | Balance | Withdrawal | Outcome           | Remaining |
	//       |    $500 |        $50 | receive $50 cash  |      $450 |
	//       |    $500 |       $100 | receive $100 cash |      $400 |
	//
	//     Examples:
	//       | Balance | Withdrawal | Outcome              | Remaining |
	//       |    $100 |       $200 | see an error message |      $100 |
	//       |      $0 |        $50 | see an error message |        $0 |
	//
}

const unformatedGherkinWithScenarioDescriptionAndBullets = `
Feature: Descriptions and Bullets
In order to be expressive as one can get
As some feature writer
I should be able to pass arbitary descriptions and organize steps with bullet points.
Scenario: Expressive as it can get
Imagine I can write lines and lines and lines of text to describe my scenario,
but what really, really counts is my the following:
* I have 6 cukes
* I eat 2 of my cukes
* I should have 4 cukes left
`

func ExampleGherkinPrettyFormater_5() {

	fmt := &formater.GherkinPrettyFormater{}

	// unformatedGherkin := `@dead @simple Feature: Dead Simple Calculator ...`
	gp := gherkin.NewGherkinDOMParser(unformatedGherkinWithScenarioDescriptionAndBullets)

	fmt.Format(gp, os.Stdout)

	// Output:
	//
	// Feature: Descriptions and Bullets
	//   In order to be expressive as one can get
	//   As some feature writer
	//   I should be able to pass arbitary descriptions and organize steps with bullet points.
	//
	//   Scenario: Expressive as it can get
	//     Imagine I can write lines and lines and lines of text to describe my scenario,
	//     but what really, really counts is my the following:
	//
	//     * I have 6 cukes
	//     * I eat 2 of my cukes
	//     * I should have 4 cukes left
	//
}
