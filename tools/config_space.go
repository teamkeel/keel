package tools

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/teamkeel/keel/casing"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

type SpaceConfigs []*SpaceConfig

// toProto will return the SpaceConfigs as protobuf messages.
func (ss SpaceConfigs) toProto() []*toolsproto.Space {
	spaces := []*toolsproto.Space{}
	for _, cfg := range ss {
		spaces = append(spaces, cfg.toProto())
	}

	return spaces
}

// findByID finds a space with the given id.
func (ss SpaceConfigs) findByID(id string) *SpaceConfig {
	for _, c := range ss {
		if c.ID == id {
			return c
		}
	}

	return nil
}

// allActions returns all actions within these spaces.
func (ss SpaceConfigs) allActions() SpaceActions {
	actions := SpaceActions{}

	for _, space := range ss {
		actions = append(actions, space.allActions()...)
	}

	return actions
}

// allGroups returns all groups within these spaces.
func (ss SpaceConfigs) allGroups() SpaceGroups {
	groups := SpaceGroups{}

	for _, space := range ss {
		groups = append(groups, space.Groups...)
	}

	return groups
}

type SpaceConfig struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Icon         string       `json:"icon"`
	DisplayOrder int32        `json:"display_order"`
	Actions      SpaceActions `json:"actions"`
	Groups       SpaceGroups  `json:"groups"`
	Links        SpaceLinks   `json:"links"`
	Metrics      SpaceMetrics `json:"metrics"`
}

// toProto will return the SpaceConfig as a protobuf message.
func (s *SpaceConfig) toProto() *toolsproto.Space {
	if s == nil {
		return nil
	}

	return &toolsproto.Space{
		Id:           s.ID,
		Name:         s.Name,
		Icon:         s.Icon,
		DisplayOrder: s.DisplayOrder,
		Actions:      s.Actions.toProto(),
		Links:        s.Links.toProto(),
		Groups:       s.Groups.toProto(),
		Metrics:      s.Metrics.toProto(),
	}
}

// setUniqueID will set a unique ID for this space config taking in account existing configs.
func (s *SpaceConfig) setUniqueID(existing SpaceConfigs) error {
	id := "space-" + casing.ToKebab(s.Name)

	if exists := existing.findByID(id); exists != nil {
		// generate a unique id suffix
		uid, err := gonanoid.Generate(nanoidABC, nanoidSize)
		if err != nil {
			return fmt.Errorf("generating unique id: %w", err)
		}
		id = id + "-" + uid
	}

	s.ID = id

	return nil
}

// allActions returns all actions within this space by taking all top level actions and appending all actions from groups.
func (s *SpaceConfig) allActions() SpaceActions {
	actions := SpaceActions{}

	actions = append(actions, s.Actions...)

	for _, group := range s.Groups {
		actions = append(actions, group.Actions...)
	}

	return actions
}

// addAction adds the given action to ths space. if a group ID is provided, the action is added within that group.
func (s *SpaceConfig) addAction(action *SpaceAction, groupID string) error {
	if groupID != "" {
		group := s.Groups.findByID(groupID)
		if group == nil {
			return fmt.Errorf("group not found")
		}

		group.Actions = append(group.Actions, action)

		return nil
	}

	s.Actions = append(s.Actions, action)

	return nil
}

type SpaceActions []*SpaceAction

func (a SpaceActions) toProto() []*toolsproto.SpaceAction {
	actions := []*toolsproto.SpaceAction{}
	for _, cfg := range a {
		actions = append(actions, &toolsproto.SpaceAction{
			Id:   cfg.ID,
			Link: cfg.Link.applyOn(nil),
		})
	}

	return actions
}

// findByID finds the action with the given id.
func (a SpaceActions) findByID(id string) *SpaceAction {
	for _, c := range a {
		if c.ID == id {
			return c
		}
	}

	return nil
}

type SpaceAction struct {
	ID   string      `json:"id"`
	Link *LinkConfig `json:"link"`
}

// setUniqueID will set a unique ID for this space action taking in account existing configs.
func (a *SpaceAction) setUniqueID(spaces SpaceConfigs) error {
	id := "action-" + casing.ToKebab(a.Link.ToolID)

	if exists := spaces.allActions().findByID(id); exists != nil {
		// generate a unique id suffix
		uid, err := gonanoid.Generate(nanoidABC, nanoidSize)
		if err != nil {
			return fmt.Errorf("generating unique id: %w", err)
		}
		id = id + "-" + uid
	}

	a.ID = id

	return nil
}

type SpaceLinks []*SpaceLink

type SpaceLink struct {
	ID   string        `json:"id"`
	Link *ExternalLink `json:"link"`
}

func (l SpaceLinks) toProto() []*toolsproto.SpaceLink {
	links := []*toolsproto.SpaceLink{}
	for _, cfg := range l {
		links = append(links, &toolsproto.SpaceLink{
			Id:   cfg.ID,
			Link: cfg.Link.toProto(),
		})
	}

	return links
}

type SpaceGroups []*SpaceGroup

func (g SpaceGroups) toProto() []*toolsproto.SpaceGroup {
	groups := []*toolsproto.SpaceGroup{}
	for _, cfg := range g {
		groups = append(groups, &toolsproto.SpaceGroup{
			Id:           cfg.ID,
			Name:         makeStringTemplate(&cfg.Name),
			Description:  makeStringTemplate(&cfg.Description),
			DisplayOrder: cfg.DisplayOrder,
			Actions:      cfg.Actions.toProto(),
		})
	}

	return groups
}

// findByID finds the group with the given id.
func (g SpaceGroups) findByID(id string) *SpaceGroup {
	for _, group := range g {
		if group.ID == id {
			return group
		}
	}

	return nil
}

type SpaceGroup struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	DisplayOrder int32        `json:"display_order"`
	Actions      SpaceActions `json:"actions"`
}

// setUniqueID will set a unique ID for this space group taking in account existing configs.
func (g *SpaceGroup) setUniqueID(spaces SpaceConfigs) error {
	id := "group-" + casing.ToKebab(g.Name)

	if exists := spaces.allGroups().findByID(id); exists != nil {
		// generate a unique id suffix
		uid, err := gonanoid.Generate(nanoidABC, nanoidSize)
		if err != nil {
			return fmt.Errorf("generating unique id: %w", err)
		}
		id = id + "-" + uid
	}

	g.ID = id

	return nil
}

type SpaceMetrics []*SpaceMetric

type SpaceMetric struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	ToolID        string `json:"tool_id"`
	FacetLocation string `json:"facet_location"`
	DisplayOrder  int32  `json:"display_order"`
}

func (m SpaceMetrics) toProto() []*toolsproto.SpaceMetric {
	metrics := []*toolsproto.SpaceMetric{}
	for _, cfg := range m {
		metrics = append(metrics, &toolsproto.SpaceMetric{
			Id:            cfg.ID,
			Label:         makeStringTemplate(&cfg.Label),
			ToolId:        cfg.ToolID,
			FacetLocation: &toolsproto.JsonPath{Path: cfg.FacetLocation},
			DisplayOrder:  cfg.DisplayOrder,
		})
	}

	return metrics
}

// findByID finds the metric with the given id.
func (m SpaceMetrics) findByID(id string) *SpaceMetric {
	for _, metric := range m {
		if metric.ID == id {
			return metric
		}
	}

	return nil
}
