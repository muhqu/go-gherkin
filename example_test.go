package gherkin_test

import (
	"fmt"

	"github.com/muhqu/go-gherkin"
)

func ExampleParseGherkinFeature() {
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

	// Output:
	// feature: "Hello World" []string{"wip"}
	// no. scenarios: 1
	// scenario 1: "Nice people" []string{"nice", "people"}
	//   step 1: "Given" "a nice person called \"Bob\""
	//   step 2: "And" "a nice person called \"Lisa\""
	//   step 3: "When" "\"Bob\" says to \"Lisa\": \"Hello!\""
	//   step 4: "Then" "\"Lisa\" should reply to \"Bob\": \"Hello!\""
}

func ExampleNewGherkinParser() {
	gp := gherkin.NewGherkinParser(`
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
	gp.WithEventProcessor(gherkin.EventProcessorFn(func(e gherkin.GherkinEvent) {
		fmt.Println(e)
	}))

	gp.Init()
	err := gp.Parse()
	if err != nil {
		panic(err)
	}
	gp.Execute()

	// Output:
	//
	// FeatureEvent("Hello World","The world is a beautiful place\nSo let people be nice to each other",["wip"])
	// ScenarioEvent("Nice people","",["nice" "people"])
	// StepEvent("Given","a nice person called \"Bob\"")
	// StepEndEvent()
	// StepEvent("And","a nice person called \"Lisa\"")
	// StepEndEvent()
	// StepEvent("When","\"Bob\" says to \"Lisa\": \"Hello!\"")
	// StepEndEvent()
	// StepEvent("Then","\"Lisa\" should reply to \"Bob\": \"Hello!\"")
	// StepEndEvent()
	// ScenarioEndEvent()
	// FeatureEndEvent()
	//
}

func ExampleNewGherkinDOMParser() {
	gp := gherkin.NewGherkinDOMParser(`
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
	feature := gp.Feature()

	fmt.Printf("feature: %#v %#v\n", feature.Title(), feature.Tags())
	fmt.Printf("no. scenarios: %#v\n", len(feature.Scenarios()))
	scenario1 := feature.Scenarios()[0]
	fmt.Printf("scenario 1: %#v %#v\n", scenario1.Title(), scenario1.Tags())

	// Output:
	// feature: "Hello World" []string{"wip"}
	// no. scenarios: 1
	// scenario 1: "Nice people" []string{"nice", "people"}
}

func ExampleLogFn() {
	gp := gherkin.NewGherkinParser(`
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

	gp.WithLogFn(func(msg string, args ...interface{}) {
		fmt.Printf(msg+"\n", args...)
	})

	gp.Init()
	err := gp.Parse()
	if err != nil {
		panic(err)
	}
	gp.Execute()

	// Output:
	// BeginFeature: "Hello World": "The world is a beautiful place\nSo let people be nice to each other" tags:[wip]
	// BeginScenario: "Nice people": "" tags:[nice people]
	// BeginStep: "Given": "a nice person called \"Bob\""
	// EndStep
	// BeginStep: "And": "a nice person called \"Lisa\""
	// EndStep
	// BeginStep: "When": "\"Bob\" says to \"Lisa\": \"Hello!\""
	// EndStep
	// BeginStep: "Then": "\"Lisa\" should reply to \"Bob\": \"Hello!\""
	// EndStep
	// EndScenario
	// EndFeature

}
