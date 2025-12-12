package tools

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/casing"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

// spaceItem is an interface constraint for types that belong to a space.
type spaceItem interface {
	GetID() string
}

var _ spaceItem = &SpaceConfig{}
var _ spaceItem = &SpaceAction{}
var _ spaceItem = &SpaceGroup{}
var _ spaceItem = &SpaceMetric{}
var _ spaceItem = &SpaceLink{}

// findByID is finds an item with the given id within the given collection.
func findByID[T spaceItem](items []T, id string) T {
	for _, item := range items {
		if item.GetID() == id {
			return item
		}
	}
	var zero T

	return zero
}

// existsByID checks if the given item exists in the given collection.
func existsByID[T spaceItem](items []T, id string) bool {
	for _, item := range items {
		if item.GetID() == id {
			return true
		}
	}

	return false
}

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
	return findByID(ss, id)
}

// allItems is a generic helper that collects items from all spaces using an extractor function.
func allItems[T any](spaces SpaceConfigs, extract func(*SpaceConfig) []T) []T {
	items := []T{}
	for _, space := range spaces {
		items = append(items, extract(space)...)
	}
	return items
}

// allActions returns all actions within these spaces.
func (ss SpaceConfigs) allActions() SpaceActions {
	return allItems(ss, func(s *SpaceConfig) []*SpaceAction { return s.allActions() })
}

// allGroups returns all groups within these spaces.
func (ss SpaceConfigs) allGroups() SpaceGroups {
	return allItems(ss, func(s *SpaceConfig) []*SpaceGroup { return s.Groups })
}

// allMetrics returns all metrics within these spaces.
func (ss SpaceConfigs) allMetrics() SpaceMetrics {
	return allItems(ss, func(s *SpaceConfig) []*SpaceMetric { return s.Metrics })
}

// allLinks returns all links within these spaces.
func (ss SpaceConfigs) allLinks() SpaceLinks {
	return allItems(ss, func(s *SpaceConfig) []*SpaceLink { return s.Links })
}

// findUniqueID will generate a unique ID for the given spaceItem. The ID will be prefixed based on the type of the item.
func (ss SpaceConfigs) findUniqueID(item spaceItem) (string, error) {
	var id string
	var exists bool

	switch tmp := item.(type) {
	case *SpaceConfig:
		id = "space-" + casing.ToKebab(tmp.Name)
		exists = existsByID(ss, id)
	case *SpaceAction:
		id = "action-" + casing.ToKebab(tmp.Link.ToolID)
		exists = existsByID(ss.allActions(), id)
	case *SpaceLink:
		id = "link-" + casing.ToKebab(tmp.Link.Label)
		exists = existsByID(ss.allLinks(), id)
	case *SpaceGroup:
		id = "group-" + casing.ToKebab(tmp.Name)
		exists = existsByID(ss.allGroups(), id)
	case *SpaceMetric:
		id = "metric-" + casing.ToKebab(tmp.Label)
		exists = existsByID(ss.allMetrics(), id)
	default:
		return "", fmt.Errorf("unknown space item type")
	}

	if exists {
		// generate a unique id suffix
		uid, err := gonanoid.Generate(nanoidABC, nanoidSize)
		if err != nil {
			return "", fmt.Errorf("generating unique id: %w", err)
		}
		id = id + "-" + uid
	}

	return id, nil
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

func (s *SpaceConfig) GetID() string {
	return s.ID
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
	id, err := existing.findUniqueID(s)
	if err != nil {
		return fmt.Errorf("finding a unique id: %w", err)
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

// removeItem will remove the item with the given id from within this space. returns true if item removed, false if not found.
func (s *SpaceConfig) removeItem(itemID string) bool {
	if item := s.Actions.findByID(itemID); item != nil {
		s.Actions = lo.Without(s.Actions, item)

		return true
	}

	if item := s.Groups.findByID(itemID); item != nil {
		s.Groups = lo.Without(s.Groups, item)

		return true
	}

	if item := s.Links.findByID(itemID); item != nil {
		s.Links = lo.Without(s.Links, item)

		return true
	}
	if item := s.Metrics.findByID(itemID); item != nil {
		s.Metrics = lo.Without(s.Metrics, item)

		return true
	}

	for _, g := range s.Groups {
		if item := g.Actions.findByID(itemID); item != nil {
			g.Actions = lo.Without(g.Actions, item)

			return true
		}
	}

	return false
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
	return findByID(a, id)
}

type SpaceAction struct {
	ID   string      `json:"id"`
	Link *LinkConfig `json:"link"`
}

func (a *SpaceAction) GetID() string {
	return a.ID
}

// setUniqueID will set a unique ID for this space action taking in account existing configs.
func (a *SpaceAction) setUniqueID(spaces SpaceConfigs) error {
	id, err := spaces.findUniqueID(a)
	if err != nil {
		return fmt.Errorf("finding a unique id: %w", err)
	}

	a.ID = id

	return nil
}

type SpaceLinks []*SpaceLink

func (ll SpaceLinks) toProto() []*toolsproto.SpaceLink {
	links := []*toolsproto.SpaceLink{}
	for _, cfg := range ll {
		links = append(links, &toolsproto.SpaceLink{
			Id:   cfg.ID,
			Link: cfg.Link.toProto(),
		})
	}

	return links
}

// findByID finds the link with the given id.
func (ll SpaceLinks) findByID(id string) *SpaceLink {
	return findByID(ll, id)
}

type SpaceLink struct {
	ID   string        `json:"id"`
	Link *ExternalLink `json:"link"`
}

func (l *SpaceLink) GetID() string {
	return l.ID
}

// setUniqueID will set a unique ID for this space link taking in account existing configs.
func (l *SpaceLink) setUniqueID(spaces SpaceConfigs) error {
	id, err := spaces.findUniqueID(l)
	if err != nil {
		return fmt.Errorf("finding a unique id: %w", err)
	}

	l.ID = id

	return nil
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
	return findByID(g, id)
}

type SpaceGroup struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	DisplayOrder int32        `json:"display_order"`
	Actions      SpaceActions `json:"actions"`
}

func (g *SpaceGroup) GetID() string {
	return g.ID
}

// setUniqueID will set a unique ID for this space group taking in account existing configs.
func (g *SpaceGroup) setUniqueID(spaces SpaceConfigs) error {
	id, err := spaces.findUniqueID(g)
	if err != nil {
		return fmt.Errorf("finding a unique id: %w", err)
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

func (m *SpaceMetric) GetID() string {
	return m.ID
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
	return findByID(m, id)
}

// setUniqueID will set a unique ID for this space group taking in account existing configs.
func (m *SpaceMetric) setUniqueID(spaces SpaceConfigs) error {
	id, err := spaces.findUniqueID(m)
	if err != nil {
		return fmt.Errorf("finding a unique id: %w", err)
	}

	m.ID = id

	return nil
}
