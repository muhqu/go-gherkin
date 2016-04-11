// Sub-Package gherkin/events provides the data-structure types for the evented gherkin parser.
package events

import (
	"fmt"
)

type EventType int

type Event interface {
	EventType() EventType
}

const (
	FeatureEventType EventType = iota
	FeatureEndEventType
	BackgroundEventType
	BackgroundEndEventType
	ScenarioEventType
	ScenarioEndEventType
	StepEventType
	StepEndEventType
	OutlineEventType
	OutlineEndEventType
	OutlineExamplesEventType
	OutlineExamplesEndEventType
	PyStringEventType
	PyStringLineEventType
	PyStringEndEventType
	TableEventType
	TableRowEventType
	TableCellEventType
	TableRowEndEventType
	TableEndEventType
	BlankLineEventType
	CommentEventType
)

type FeatureEvent struct {
	Title       string
	Description string
	Tags        []string
}

func (*FeatureEvent) EventType() EventType {
	return FeatureEventType
}
func (e *FeatureEvent) String() string {
	return fmt.Sprintf("FeatureEvent(%q,%q,%q)", e.Title, e.Description, e.Tags)
}

type FeatureEndEvent struct {
}

func (*FeatureEndEvent) EventType() EventType {
	return FeatureEndEventType
}
func (*FeatureEndEvent) String() string {
	return "FeatureEndEvent()"
}

type BackgroundEvent struct {
	Title       string
	Description string
	Tags        []string
}

func (*BackgroundEvent) EventType() EventType {
	return BackgroundEventType
}
func (e *BackgroundEvent) String() string {
	return fmt.Sprintf("BackgroundEvent(%q,%q,%q)", e.Title, e.Description, e.Tags)
}

type BackgroundEndEvent struct {
}

func (*BackgroundEndEvent) EventType() EventType {
	return BackgroundEndEventType
}
func (*BackgroundEndEvent) String() string {
	return "BackgroundEndEvent()"
}

type ScenarioEvent struct {
	Title       string
	Description string
	Tags        []string
}

func (*ScenarioEvent) EventType() EventType {
	return ScenarioEventType
}
func (e *ScenarioEvent) String() string {
	return fmt.Sprintf("ScenarioEvent(%q,%q,%q)", e.Title, e.Description, e.Tags)
}

type ScenarioEndEvent struct {
}

func (*ScenarioEndEvent) EventType() EventType {
	return ScenarioEndEventType
}
func (*ScenarioEndEvent) String() string {
	return "ScenarioEndEvent()"
}

type OutlineEvent struct {
	Title       string
	Description string
	Tags        []string
}

func (e *OutlineEvent) EventType() EventType {
	return OutlineEventType
}
func (e *OutlineEvent) String() string {
	return fmt.Sprintf("OutlineEvent(%q,%q,%q)", e.Title, e.Description, e.Tags)
}

type OutlineEndEvent struct {
}

func (*OutlineEndEvent) EventType() EventType {
	return OutlineEndEventType
}
func (*OutlineEndEvent) String() string {
	return "OutlineEndEvent()"
}

type OutlineExamplesEvent struct {
	Title string
}

func (*OutlineExamplesEvent) EventType() EventType {
	return OutlineExamplesEventType
}
func (*OutlineExamplesEvent) String() string {
	return "OutlineExamplesEvent()"
}

type OutlineExamplesEndEvent struct {
}

func (*OutlineExamplesEndEvent) EventType() EventType {
	return OutlineExamplesEndEventType
}
func (*OutlineExamplesEndEvent) String() string {
	return "OutlineExamplesEndEvent()"
}

type StepEvent struct {
	StepType string
	Text     string
}

func (*StepEvent) EventType() EventType {
	return StepEventType
}
func (e *StepEvent) String() string {
	return fmt.Sprintf("StepEvent(%q,%q)", e.StepType, e.Text)
}

type StepEndEvent struct {
}

func (*StepEndEvent) EventType() EventType {
	return StepEndEventType
}
func (*StepEndEvent) String() string {
	return "StepEndEvent()"
}

type PyStringEvent struct {
	Intent string
}

func (*PyStringEvent) EventType() EventType {
	return PyStringEventType
}
func (e *PyStringEvent) String() string {
	return fmt.Sprintf("PyStringEvent(%q)", e.Intent)
}

type PyStringLineEvent struct {
	Line string
}

func (*PyStringLineEvent) EventType() EventType {
	return PyStringLineEventType
}
func (e *PyStringLineEvent) String() string {
	return fmt.Sprintf("PyStringLineEvent(%q)", e.Line)
}

type PyStringEndEvent struct {
}

func (*PyStringEndEvent) EventType() EventType {
	return PyStringEndEventType
}
func (*PyStringEndEvent) String() string {
	return "PyStringEndEvent()"
}

type TableEvent struct {
}

func (*TableEvent) EventType() EventType {
	return TableEventType
}
func (*TableEvent) String() string {
	return "TableEvent()"
}

type TableRowEvent struct {
}

func (*TableRowEvent) EventType() EventType {
	return TableRowEventType
}
func (*TableRowEvent) String() string {
	return "TableRowEvent()"
}

type TableRowEndEvent struct {
}

func (*TableRowEndEvent) EventType() EventType {
	return TableRowEndEventType
}
func (*TableRowEndEvent) String() string {
	return "TableRowEndEvent()"
}

type TableCellEvent struct {
	Content string
}

func (*TableCellEvent) EventType() EventType {
	return TableCellEventType
}
func (e *TableCellEvent) String() string {
	return fmt.Sprintf("TableCellEvent(%q)", e.Content)
}

type TableEndEvent struct {
}

func (*TableEndEvent) EventType() EventType {
	return TableEndEventType
}
func (*TableEndEvent) String() string {
	return "TableEndEvent()"
}

type BlankLineEvent struct {
}

func (*BlankLineEvent) EventType() EventType {
	return BlankLineEventType
}
func (*BlankLineEvent) String() string {
	return "BlankLineEvent()"
}

type CommentEvent struct {
	Comment string
}

func (*CommentEvent) EventType() EventType {
	return CommentEventType
}
func (e *CommentEvent) String() string {
	return fmt.Sprintf("CommentEvent(%q)", e.Comment)
}
