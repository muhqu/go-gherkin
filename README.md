Go Gherkin package
======================

[![Build Status](https://travis-ci.org/muhqu/go-gherkin.png?branch=master)](https://travis-ci.org/muhqu/go-gherkin)

Package gherkin implements a parser for the Gherkin language, used for
Story/Feature based Behavior Driven Development.

The parser is generated via [@pointlander]'s awesome [peg] parser generator. 

Read the [Documentation][godoc] over at [godoc.org][godoc].

Install
-------

### Via Go Get

```bash
$ go get github.com/muhqu/go-gherkin
```

### Via GNU Make

```bash
$ git clone https://github.com/muhqu/go-gherkin
...
$ cd go-gherkin/
$ make get-deps build test install
```


Usage Example
-------------
```go
import (
  "github.com/muhqu/go-gherkin"
)

func main() {
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
}
```

Output:
```
feature: "Hello World" []string{"wip"}
no. scenarios: 1
scenario 1: "Nice people" []string{"nice", "people"}
  step 1: "Given" "a nice person called \"Bob\""
  step 2: "And" "a nice person called \"Lisa\""
  step 3: "When" "\"Bob\" says to \"Lisa\": \"Hello!\""
  step 4: "Then" "\"Lisa\" should reply to \"Bob\": \"Hello!\""
```

Author
------

|   |   |
|---|---|
| ![](http://gravatar.com/avatar/0ad964bc2b83e0977d8f70816eda1c70) | © 2014 by Mathias Leppich <br>  [github.com/muhqu](https://github.com/muhqu), [@muhqu](http://twitter.com/muhqu) |
|   |   |

[godoc]: http://godoc.org/github.com/muhqu/go-gherkin
[@pointlander]: http://github.com/pointlander
[peg]: http://github.com/pointlander/peg
