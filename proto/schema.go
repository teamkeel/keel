package proto

import (
	"sort"

	"github.com/samber/lo"
)

// HasFiles checks if the given schema has any models with fields that are files
func (p *Schema) HasFiles() bool {
	for _, model := range p.Models {
		if model.HasFiles() {
			return true
		}
	}

	return false
}

// ModelNames provides a (sorted) list of all the Model names used in the
// given schema.
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

func (s *Schema) FindAction(actionName string) *Action {
	actions := s.FilterActions(func(op *Action) bool {
		return op.Name == actionName
	})
	if len(actions) != 1 {
		return nil
	}
	return actions[0]
}

// FindEventSubscribers locates the subscribers for the given event.
func (s *Schema) FindEventSubscribers(event *Event) []*Subscriber {
	subscribers := lo.Filter(s.Subscribers, func(m *Subscriber, _ int) bool {
		return lo.Contains(m.EventNames, event.Name)
	})
	return subscribers
}
