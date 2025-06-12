package flows

import (
	"encoding/json"
	"fmt"
)

const (
	EventNameFlowRunUpdated = "flow.updated"
	EventNameFlowRunStarted = "flow.started"
)

type EventWrapper struct {
	// The name of the event, e.g. flow.updated
	EventName string `json:"eventName"`
	// Payload of event
	Payload     string `json:"payload"`
	Traceparent string `json:"traceparent,omitempty"`
}

type FlowRunStarted struct {
	// The name of the flow to run e.g. MySpecialFlow
	Name   string         `json:"name"`
	Inputs map[string]any `json:"inputs"`
}

func (e *FlowRunStarted) ReadPayload(ev *EventWrapper) error {
	if ev == nil {
		return fmt.Errorf("invalid event ")
	}

	if ev.EventName != EventNameFlowRunStarted {
		return fmt.Errorf("invalid event type")
	}

	return json.Unmarshal([]byte(ev.Payload), e)
}

func (e *FlowRunStarted) Wrap() (*EventWrapper, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return &EventWrapper{
		EventName: EventNameFlowRunStarted,
		Payload:   string(payload),
	}, nil
}

type FlowRunUpdated struct {
	RunID  string         `json:"runId"`
	Data   map[string]any `json:"data"`
	Action string         `json:"action,omitempty"`
}

func (e *FlowRunUpdated) ReadPayload(ev *EventWrapper) error {
	if ev == nil {
		return fmt.Errorf("invalid event ")
	}

	if ev.EventName != EventNameFlowRunUpdated {
		return fmt.Errorf("invalid event type")
	}

	return json.Unmarshal([]byte(ev.Payload), e)
}

func (e *FlowRunUpdated) Wrap() (*EventWrapper, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	return &EventWrapper{
		EventName: EventNameFlowRunUpdated,
		Payload:   string(payload),
	}, nil
}
