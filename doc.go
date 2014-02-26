/*
Package gherkin implements a parser for the Gherkin language, used for
Story/Feature based Behavior Driven Development.

Basic usage example:

	feature, _ := gherkin.ParseGherkinFeature(`
	@wip
	Feature: Hello World
	  The world is a beautiful place
	  So let people be nice to each other

	  @nice @people
	  Scenario: Nice people
	    Given a nice person called "Bob"
	      And a nice person called "Lisa"
	     When "Bob" says to "Lisa": "Hello!"
	     Then "Lisa" should reply to "Bob": "Hello!"

	`)

	fmt.Printf("feature: %#v %#v\n", feature.Title(), feature.Tags())
	fmt.Printf("no. scenarios: %#v\n", len(feature.Scenarios()))
	for i, scenario := range feature.Scenarios() {
		fmt.Printf("scenario %d: %#v %#v\n", i+1, scenario.Title(), scenario.Tags())
		for i, step := range scenario.Steps() {
			fmt.Printf("  step %d: %#v %#v\n", i+1, step.StepType(), step.Text())
		}
	}

Output:
	feature: "Hello World" []string{"wip"}
	no. scenarios: 1
	scenario 1: "Nice people" []string{"nice", "people"}
	  step 1: "Given" "a nice person called \"Bob\""
	  step 2: "And" "a nice person called \"Lisa\""
	  step 3: "When" "\"Bob\" says to \"Lisa\": \"Hello!\""
	  step 4: "Then" "\"Lisa\" should reply to \"Bob\": \"Hello!\""

*/
package gherkin
