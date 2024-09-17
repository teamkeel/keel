package proto

import (
	"sort"

	"github.com/samber/lo"
)

// HasFiles checks if the given schema has any models or messages with fields that are files
func (s *Schema) HasFiles() bool {
	for _, model := range s.Models {
		if model.HasFiles() {
			return true
		}
	}

	for _, message := range s.Messages {
		if messageHasFiles(s, message) {
			return true
		}
	}

	return false
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
	for _, m := range s.Models {
		for _, a := range m.GetActions() {
			if a.InputMessageName == messageName {
				return true
			}

			msg := s.FindMessage(a.InputMessageName)
			if msg.hasMessage(s, messageName) {
				return true
			}
		}
	}
	return false
}

func (m *Message) hasMessage(s *Schema, messageName string) bool {
	for _, f := range m.Fields {
		if f.Type.Type == Type_TYPE_MESSAGE {
			if f.Type.MessageName.Value == messageName {
				return true
			}

			msg := s.FindMessage(f.Type.MessageName.Value)
			if msg.hasMessage(s, messageName) {
				return true
			}
		}
	}
	return false
}

// IsActionResponseMessage returns true if the message is used to define an action's response.
func (s *Schema) IsActionResponseMessage(messageName string) bool {
	for _, m := range s.Models {
		for _, a := range m.GetActions() {
			if a.ResponseMessageName == messageName {
				return true
			}

			if a.ResponseMessageName != "" {
				msg := s.FindMessage(a.ResponseMessageName)
				if msg.hasMessage(s, messageName) {
					return true
				}
			}

		}
	}
	return false
}

// ModelNames provides a (sorted) list of all the Model names used in the given schema.
func (s *Schema) ModelNames() []string {
	names := lo.Map(s.Models, func(x *Model, _ int) string {
		return x.Name
	})
	sort.Strings(names)
	return names
}

// AllFields provides a list of all the model fields specified in the schema.
func (s *Schema) AllFields() []*Field {
	fields := []*Field{}
	for _, model := range s.Models {
		fields = append(fields, model.Fields...)
	}
	return fields
}

func (s *Schema) FilterActions(filter func(op *Action) bool) (ops []*Action) {
	for _, model := range s.Models {
		actions := model.Actions

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
		return op.Name == actionName
	})
	if len(actions) != 1 {
		return nil
	}
	return actions[0]
}

// FindJob locates the job of the given name.
func (s *Schema) FindJob(name string) *Job {
	job, _ := lo.Find(s.Jobs, func(m *Job) bool {
		return m.Name == name
	})
	return job
}

// FindEventSubscribers locates the subscribers for the given event.
func (s *Schema) FindEventSubscribers(event *Event) []*Subscriber {
	subscribers := lo.Filter(s.Subscribers, func(m *Subscriber, _ int) bool {
		return lo.Contains(m.EventNames, event.Name)
	})
	return subscribers
}

// FindApiNames finds the api name for the given model and action name.
func (s *Schema) FindApiNames(modelName, actionName string) []string {
	names := []string{}

	for _, api := range s.Apis {
		for _, apiModel := range api.ApiModels {
			if apiModel.ModelName == modelName {
				for _, action := range apiModel.ModelActions {
					if action.ActionName == actionName {
						names = append(names, api.Name)
					}
				}
			}
		}
	}

	return names
}
