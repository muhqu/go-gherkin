package formater_test

import (
	"fmt"
	"os"

	"github.com/muhqu/go-gherkin/formater"
	"github.com/muhqu/go-gherkin/nodes"
)

func ExampleGherkinPrettyFormater_FormatStep_stepWithTable() {
	fmt.Println(">")

	gfmt := &formater.GherkinPrettyFormater{}

	step := nodes.NewMutableStepNode("Given", "the following users:").
		WithTable(nodes.NewMutableTableNode().WithRows([][]string{
		{"username", "email"},
		{"Foobar", "foo@bar.org"},
		{"JohnDoe", "naked-john74@hotmail.com"},
	}))

	gfmt.FormatStep(step, os.Stdout)

	fmt.Println(">")
	// Output:
	// >
	//     Given the following users:
	//       | username | email                    |
	//       | Foobar   | foo@bar.org              |
	//       | JohnDoe  | naked-john74@hotmail.com |
	// >
}

func ExampleGherkinPrettyFormater_FormatStep_stepWithPyString() {
	fmt.Println(">")

	gfmt := &formater.GherkinPrettyFormater{}

	step := nodes.NewMutableStepNode("Given", "the following user relations:").
		WithPyString(nodes.NewMutablePyStringNode().WithLines([]string{
		"Jenny [follows] Mary, David",
		"Bill [knows] Mary, Jenny, David",
	}))

	gfmt.FormatStep(step, os.Stdout)

	fmt.Println(">")
	// Output:
	// >
	//     Given the following user relations:
	//       """
	//       Jenny [follows] Mary, David
	//       Bill [knows] Mary, Jenny, David
	//       """
	// >
}

func ExampleGherkinPrettyFormater_FormatStep_givenWhenThen() {
	fmt.Println(">")

	gfmt := &formater.GherkinPrettyFormater{CenterSteps: true}

	step := nodes.NewMutableStepNode("Given", "I have 2 banannas")
	gfmt.FormatStep(step, os.Stdout)

	step = nodes.NewMutableStepNode("When", "I eat 1 bananna")
	gfmt.FormatStep(step, os.Stdout)

	step = nodes.NewMutableStepNode("And", "I throw 1 bananna away")
	gfmt.FormatStep(step, os.Stdout)

	step = nodes.NewMutableStepNode("Then", "I should still have 2 banannas")
	gfmt.FormatStep(step, os.Stdout)

	fmt.Println(">")
	// Output:
	// >
	//     Given I have 2 banannas
	//      When I eat 1 bananna
	//       And I throw 1 bananna away
	//      Then I should still have 2 banannas
	// >
}

func ExampleGherkinPrettyFormater_FormatStep_givenWhenThenComments() {
	fmt.Println(">")

	gfmt := &formater.GherkinPrettyFormater{CenterSteps: true, AnsiColors: false}

	scenario := nodes.NewMutableScenarioNode("Awesome", nil)
	scenario.SetComment(nodes.NewCommentNode("scenario comment"))
	step := nodes.NewMutableStepNode("Given", "I have 2 banannas")
	step.SetComment(nodes.NewCommentNode("first step comment"))
	scenario.AddStep(step)

	step = nodes.NewMutableStepNode("When", "I eat 1 bananna")
	step.SetComment(nodes.NewCommentNode("2nd step comment"))
	scenario.AddStep(step)

	step = nodes.NewMutableStepNode("And", "I throw 1 bananna away")
	step.SetComment(nodes.NewCommentNode("3rd step comment"))
	scenario.AddStep(step)

	step = nodes.NewMutableStepNode("Then", "I should still have 2 banannas")
	step.SetComment(nodes.NewCommentNode("4th step comment"))
	scenario.AddStep(step)
	gfmt.FormatScenario(scenario, os.Stdout)

	fmt.Println(">")
	// Output:
	// >
	//   Scenario: Awesome                           # scenario comment
	//     Given I have 2 banannas                   # first step comment
	//      When I eat 1 bananna                     # 2nd step comment
	//       And I throw 1 bananna away              # 3rd step comment
	//      Then I should still have 2 banannas      # 4th step comment
	// >
}
