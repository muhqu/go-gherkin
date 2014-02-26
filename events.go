package gherkin

import (
	"fmt"
)

type NodeEventProcessor interface {
	ProcessNodeEvent(NodeEvent)
}

type ProcessNodeEvent func(NodeEvent)

func (fn ProcessNodeEvent) ProcessNodeEvent(e NodeEvent) {
	fn(e)
}

type EventType int

const (
	BeginNodeEventType EventType = iota
	EndNodeEventType
)

func (et EventType) String() string {
	switch et {
	case BeginNodeEventType:
		return "BeginNode"
	case EndNodeEventType:
		return "EndNode"
	}
	return "Unknown"
}

type NodeEvent interface {
	EventType() EventType
	Node() NodeInterface
	String() string
}

type nodeEvent struct {
	eventType EventType
	node      NodeInterface
}

func (n *nodeEvent) EventType() EventType {
	return n.eventType
}

func (n *nodeEvent) Node() NodeInterface {
	return n.node
}

func (n *nodeEvent) String() string {
	return fmt.Sprintf("%s(%s)", n.EventType().String(), n.Node().NodeType().String())
}

func BeginNodeEvent(node NodeInterface) NodeEvent {
	return &nodeEvent{
		eventType: BeginNodeEventType,
		node:      node,
	}
}
func EndNodeEvent(node NodeInterface) NodeEvent {
	return &nodeEvent{
		eventType: EndNodeEventType,
		node:      node,
	}
}
