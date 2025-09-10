package proto

import (
	"sort"
	"strings"

	"github.com/samber/lo"
)

// HasFiles checks if the given schema has any models or messages with fields that are files.
func (s *Schema) HasFiles() bool {
	for _, model := range s.GetModels() {
		if model.HasFiles() {
			return true
		}
	}

	for _, message := range s.GetMessages() {
		if messageHasFiles(s, message) {
			return true
		}
	}

	return false
}

// FindEntity finds within the schema the model or task that has the given name. Returns nil if model not found.
func (s *Schema) FindEntity(entityName string) Entity {
	for _, e := range s.Entities() {
		if e.GetName() == entityName {
			return e
		}
	}

	return nil
}

// FindModel finds within the schema the model that has the given name. Returns nil if model not found.
func (s *Schema) FindModel(modelName string) *Model {
	for _, m := range s.GetModels() {
		if m.GetName() == modelName {
			return m
		}
	}

	return nil
}

// FindMessage finds within the schema the message that has the given name. Returns nil if message not found.
func (s *Schema) FindMessage(messageName string) *Message {
	for _, m := range s.GetMessages() {
		if m.GetName() == messageName {
			return m
		}
	}

	return nil
}

// IsActionInputMessage returns true if the message is used to define an action's inputs.
func (s *Schema) IsActionInputMessage(messageName string) bool {
	for _, m := range s.GetModels() {
		for _, a := range m.GetActions() {
			if a.GetInputMessageName() == messageName {
				return true
			}

			msg := s.FindMessage(a.GetInputMessageName())
			if msg.hasMessage(s, messageName) {
				return true
			}
		}
	}
	return false
}

// IsActionResponseMessage returns true if the message is used to define an action's response.
func (s *Schema) IsActionResponseMessage(messageName string) bool {
	for _, m := range s.GetModels() {
		for _, a := range m.GetActions() {
			if a.GetResponseMessageName() == messageName {
				return true
			}

			if a.GetResponseMessageName() != "" {
				msg := s.FindMessage(a.GetResponseMessageName())
				if msg.hasMessage(s, messageName) {
					return true
				}
			}
		}
	}
	return false
}

// hasMessage will check to see if a message has a field of type messageName recusively.
func (m *Message) hasMessage(s *Schema, messageName string) bool {
	for _, f := range m.GetFields() {
		if f.GetType().GetType() == Type_TYPE_MESSAGE {
			if f.GetType().GetMessageName().GetValue() == messageName {
				return true
			}

			msg := s.FindMessage(f.GetType().GetMessageName().GetValue())
			if msg.hasMessage(s, messageName) {
				return true
			}
		}
	}
	return false
}

// ModelNames provides a (sorted) list of all the Model names used in the given schema.
func (s *Schema) ModelNames() []string {
	names := lo.Map(s.GetModels(), func(x *Model, _ int) string {
		return x.GetName()
	})
	sort.Strings(names)
	return names
}

// TaskNames provides a (sorted) list of all the Task names used in the given schema.
func (s *Schema) TaskNames() []string {
	names := lo.Map(s.GetTasks(), func(x *Task, _ int) string {
		return x.GetName()
	})
	sort.Strings(names)
	return names
}

// AllFields provides a list of all the model fields specified in the schema.
func (s *Schema) AllFields() []*Field {
	fields := []*Field{}
	for _, model := range s.GetModels() {
		fields = append(fields, model.GetFields()...)
	}
	return fields
}

func (s *Schema) FilterActions(filter func(op *Action) bool) (ops []*Action) {
	for _, model := range s.GetModels() {
		actions := model.GetActions()

		for _, o := range actions {
			if filter(o) {
				ops = append(ops, o)
			}
		}
	}

	return ops
}

// FindAction finds the action with the given name. Returns nil if action is not found.
func (s *Schema) FindAction(actionName string) *Action {
	actions := s.FilterActions(func(op *Action) bool {
		return op.GetName() == actionName
	})
	if len(actions) != 1 {
		return nil
	}
	return actions[0]
}

// FindFlow finds the flow with the given name. Returns nil if a flow is not found. The matching is case insensitive.
func (s *Schema) FindFlow(flowName string) *Flow {
	for _, f := range s.GetFlows() {
		if strings.EqualFold(f.GetName(), flowName) {
			return f
		}
	}

	return nil
}

// FindEnum finds within the schema the enum that has the given name. Returns nil if enum not found.
func (s *Schema) FindEnum(enumName string) *Enum {
	for _, e := range s.GetEnums() {
		if e.GetName() == enumName {
			return e
		}
	}

	return nil
}

// FindJob locates the job of the given name.
func (s *Schema) FindJob(name string) *Job {
	job, _ := lo.Find(s.GetJobs(), func(m *Job) bool {
		return m.GetName() == name
	})

	return job
}

// FindTask finds the task with the given name. Returns nil if a task is not found. The matching is case insensitive.
func (s *Schema) FindTask(name string) *Task {
	for _, t := range s.GetTasks() {
		if strings.EqualFold(t.GetName(), name) {
			return t
		}
	}

	return nil
}

// FindEventSubscribers locates the subscribers for the given event.
func (s *Schema) FindEventSubscribers(event *Event) []*Subscriber {
	subscribers := lo.Filter(s.GetSubscribers(), func(m *Subscriber, _ int) bool {
		return lo.Contains(m.GetEventNames(), event.GetName())
	})
	return subscribers
}

// FindApiNames finds the api name for the given model and action name.
func (s *Schema) FindApiNames(modelName, actionName string) []string {
	names := []string{}

	for _, api := range s.GetApis() {
		for _, apiModel := range api.GetApiModels() {
			if apiModel.GetModelName() == modelName {
				for _, action := range apiModel.GetModelActions() {
					if action.GetActionName() == actionName {
						names = append(names, api.GetName())
					}
				}
			}
		}
	}

	return names
}

// FlowNames returns an array with the names of all flows defined in this schema.
func (s *Schema) FlowNames() []string {
	names := []string{}
	for _, f := range s.GetFlows() {
		names = append(names, f.GetName())
	}

	return names
}

// ScheduledFlowNames returns an array with the names of all scheduled flows defined in this schema.
func (s *Schema) ScheduledFlowNames() []string {
	names := []string{}
	for _, f := range s.ScheduledFlows() {
		names = append(names, f.GetName())
	}

	return names
}

// HasFlows indicates if the schema has any flows defined.
func (s *Schema) HasFlows() bool {
	return len(s.GetFlows()) > 0
}

// HasScheduledFlows checks if there are any scheduled flows defined.
func (s *Schema) HasScheduledFlows() bool {
	for _, f := range s.GetFlows() {
		if f.GetSchedule() != nil {
			return true
		}
	}

	return false
}

// ScheduledFlows returns a slice of Flows that have schedules defined.
func (s *Schema) ScheduledFlows() []*Flow {
	flows := []*Flow{}

	for _, f := range s.GetFlows() {
		if f.GetSchedule() != nil {
			flows = append(flows, f)
		}
	}

	return flows
}

// GetFlowModelInputs returns a map of model names that are used as inputs to the given flow. The boolean
// value represents whether the input is required or not.
func (s *Schema) GetFlowModelInputs(f *Flow) map[string]bool {
	if f == nil || f.GetInputMessageName() == "" {
		return nil
	}
	// get the input message
	msg := s.FindMessage(f.GetInputMessageName())
	if msg == nil {
		return nil
	}

	m := map[string]bool{}

	for _, f := range msg.GetFields() {
		if mName := f.GetType().GetEntityName().GetValue(); mName != "" {
			m[mName] = !f.GetOptional() || m[mName]
		}
	}

	return m
}
