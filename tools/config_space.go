package tools

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/teamkeel/keel/casing"
	toolsproto "github.com/teamkeel/keel/tools/proto"
)

type SpaceConfigs []*SpaceConfig

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

// toProto will return the SpaceConfigs as protobuf messages.
func (s SpaceConfigs) toProto() []*toolsproto.Space {
	spaces := []*toolsproto.Space{}
	for _, cfg := range s {
		spaces = append(spaces, cfg.toProto())
	}

	return spaces
}

// findByID finds a space with the given id
func (s SpaceConfigs) findByID(id string) *SpaceConfig {
	for _, c := range s {
		if c.ID == id {
			return c
		}
	}

	return nil
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

// setUniqueID will set a unique ID for this space config taking in account exitsing configs
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

type SpaceActions []*SpaceAction

type SpaceAction struct {
	ID   string     `json:"id"`
	Link LinkConfig `json:"link"`
}

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

type SpaceLinks []*SpaceLink

type SpaceLink struct {
	ID   string       `json:"id"`
	Link ExternalLink `json:"link"`
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

type SpaceGroup struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	DisplayOrder int32        `json:"display_order"`
	Actions      SpaceActions `json:"actions"`
}

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
