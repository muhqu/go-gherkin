package gherkin_test

import (
	"github.com/muhqu/go-gherkin"
	"os"
)

func ExampleGherkinDOMWriter() {

	gp := gherkin.NewGherkinDOMParser(`@dead @simple Feature: Dead Simple Calculator         
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

	gp.Write(os.Stdout)
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
	//       | left | operator | right | result |
	//       |    2 | +        |     2 |      4 |
	//       |    3 | +        |     4 |      7 |
	//
}
