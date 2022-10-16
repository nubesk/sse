package sse

import (
	"fmt"
	"strings"
)

type Event struct {
	Name string
	Data string
}

func NewEventFromString(s string) *Event {
	e := new(Event)
	ls := strings.Split(s, "\n")
	for _, l := range ls {
		if strings.Contains(l, "event") {
			e.Name = l[7:]
		}
		if strings.Contains(l, "data") {
			e.Data = l[6:]
		}
	}

	return e
}

func NewEvent(name string, data string) *Event {
	return &Event{
		Name: name,
		Data: data,
	}
}

func (e *Event) String() string {
	b := strings.Builder{}
	b.WriteString(fmt.Sprintf("event: %s\n", e.Name))
	b.WriteString(fmt.Sprintf("data: %s\n", e.Data))
	b.WriteString("\n")
	return b.String()
}
